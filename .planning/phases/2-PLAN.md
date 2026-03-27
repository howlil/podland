# Phase 2 Plan: Core VM

**Phase:** 2 of 5
**Goal:** Users can create and manage VMs with resource quotas enforced
**Requirements:** 16 (VM-01 through VM-11, QUOTA-01 through QUOTA-05, API-01 through API-04)
**Duration:** 4 weeks
**Status:** Ready for implementation

---

## Success Criteria

1. ✅ User can create VM with name, OS, and tier (nano → xlarge)
2. ✅ VM appears in list with status "running" within 30 seconds
3. ✅ User can stop VM, status changes to "stopped"
4. ✅ User can start stopped VM, status changes to "running"
5. ✅ User can restart VM (stop → start sequence)
6. ✅ User can delete VM with confirmation dialog
7. ✅ VM list shows all user's VMs with status badges
8. ✅ VM detail shows resource usage, domain, created date
9. ✅ External user cannot create VM exceeding 0.5 CPU / 1GB RAM
10. ✅ User with full quota sees "Quota exceeded" error on create
11. ✅ VM namespace has NetworkPolicy denying inter-namespace traffic
12. ✅ VM container runs as UID 1000, not root
13. ✅ Dashboard shows quota usage bar (CPU, RAM, storage)
14. ✅ Superadmin can change user quota via database
15. ✅ API endpoint POST /api/vms creates VM with valid JWT
16. ✅ API request without JWT returns 401

---

## Technical Milestones

- [ ] Database schema migration (vms, user_quotas, tiers, user_quota_usage)
- [ ] Backend k8s module (VMManager with client-go)
- [ ] Backend VM handlers (CRUD API endpoints)
- [ ] Backend quota enforcement (SELECT FOR UPDATE, validation)
- [ ] Backend SSH key generation (Ed25519)
- [ ] Frontend VM list view (table with status badges)
- [ ] Frontend VM detail page (resource usage, SSH key display)
- [ ] Frontend Create VM wizard (tier selection, OS selection)
- [ ] k3s Traefik configuration (SSH ingress, wildcard HTTP)
- [ ] GitHub Actions workflow (VM image builds)
- [ ] Integration testing (end-to-end VM lifecycle)
- [ ] Load testing (concurrent VM creation, quota race conditions)

---

## Implementation Tasks

### Week 1: Database + Backend Core

#### Task 1.1: Database Schema Migration
**Estimate:** 4 hours
**Acceptance Criteria:**
- [ ] vms table created with all required fields
- [ ] user_quotas table created with default values
- [ ] tiers table created with 6 tiers (nano → xlarge)
- [ ] user_quota_usage table created for tracking
- [ ] Migration script tested on local PostgreSQL

**Implementation:**
```sql
-- migrations/002_phase2_vm_quota.sql

-- vms table
CREATE TABLE vms (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  name VARCHAR(100) NOT NULL,
  os VARCHAR(50) NOT NULL DEFAULT 'ubuntu-2204',
  tier VARCHAR(20) NOT NULL,
  cpu DECIMAL(4,2) NOT NULL,
  ram BIGINT NOT NULL,
  storage BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  k8s_namespace VARCHAR(100),
  k8s_deployment VARCHAR(100),
  k8s_service VARCHAR(100),
  k8s_pvc VARCHAR(100),
  domain VARCHAR(255),
  ssh_public_key TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  started_at TIMESTAMP,
  stopped_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX idx_vms_user_id ON vms(user_id);
CREATE INDEX idx_vms_status ON vms(status);

-- user_quotas table
CREATE TABLE user_quotas (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  cpu_limit DECIMAL(4,2) NOT NULL DEFAULT 0.5,
  ram_limit BIGINT NOT NULL DEFAULT 1073741824,
  storage_limit BIGINT NOT NULL DEFAULT 10737418240,
  vm_count_limit INTEGER NOT NULL DEFAULT 2,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- user_quota_usage table
CREATE TABLE user_quota_usage (
  user_id UUID PRIMARY KEY REFERENCES users(id),
  cpu_used DECIMAL(4,2) NOT NULL DEFAULT 0,
  ram_used BIGINT NOT NULL DEFAULT 0,
  storage_used BIGINT NOT NULL DEFAULT 0,
  vm_count INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- tiers table
CREATE TABLE tiers (
  name VARCHAR(20) PRIMARY KEY,
  cpu DECIMAL(4,2) NOT NULL,
  ram BIGINT NOT NULL,
  storage BIGINT NOT NULL,
  min_role VARCHAR(20) NOT NULL DEFAULT 'external'
);

-- Insert default tiers
INSERT INTO tiers (name, cpu, ram, storage, min_role) VALUES
  ('nano', 0.25, 536870912, 5368709120, 'external'),
  ('micro', 0.5, 1073741824, 10737418240, 'external'),
  ('small', 1.0, 2147483648, 21474836480, 'internal'),
  ('medium', 2.0, 4294967296, 42949672960, 'internal'),
  ('large', 4.0, 8589934592, 85899345920, 'internal'),
  ('xlarge', 4.0, 8589934592, 107374182400, 'internal');

-- Trigger: Auto-create quota on user creation
CREATE OR REPLACE FUNCTION create_user_quota()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO user_quotas (user_id, cpu_limit, ram_limit, storage_limit, vm_count_limit)
  VALUES (
    NEW.id,
    CASE WHEN NEW.nim LIKE '%1152%' THEN 4.0 ELSE 0.5 END,
    CASE WHEN NEW.nim LIKE '%1152%' THEN 8589934592 ELSE 1073741824 END,
    CASE WHEN NEW.nim LIKE '%1152%' THEN 107374182400 ELSE 10737418240 END,
    CASE WHEN NEW.nim LIKE '%1152%' THEN 5 ELSE 2 END
  );
  
  INSERT INTO user_quota_usage (user_id, cpu_used, ram_used, storage_used, vm_count)
  VALUES (NEW.id, 0, 0, 0, 0);
  
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_user_quota
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_user_quota();
```

---

#### Task 1.2: Backend k8s Module (VMManager)
**Estimate:** 8 hours
**Acceptance Criteria:**
- [ ] VMManager struct created with client-go initialization
- [ ] CreateVM method creates namespace, PVC, Deployment, Service, Ingress
- [ ] DeleteVM method deletes all k8s resources
- [ ] GetVMStatus method queries k8s for pod status
- [ ] StartVM/StopVM methods scale Deployment replicas
- [ ] Error handling for quota exceeded, name collision, etc.

**Implementation:**
```go
// backend/internal/k8s/vm_manager.go
package k8s

import (
    "context"
    "fmt"
    
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    
    "github.com/podland/backend/internal/database/types"
)

type VMManager struct {
    clientset *kubernetes.Clientset
}

func NewVMManager(kubeconfig string) (*VMManager, error) {
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
    }
    
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %w", err)
    }
    
    return &VMManager{clientset: clientset}, nil
}

// CreateVM creates all k8s resources for a VM
func (m *VMManager) CreateVM(ctx context.Context, vm *types.VM) error {
    namespace := fmt.Sprintf("user-%s", vm.UserID)
    
    // 1. Ensure namespace exists
    err := m.ensureNamespace(ctx, namespace)
    if err != nil {
        return fmt.Errorf("failed to ensure namespace: %w", err)
    }
    
    // 2. Create PVC
    err = m.createPVC(ctx, namespace, vm)
    if err != nil {
        return fmt.Errorf("failed to create PVC: %w", err)
    }
    
    // 3. Create Deployment
    err = m.createDeployment(ctx, namespace, vm)
    if err != nil {
        return fmt.Errorf("failed to create Deployment: %w", err)
    }
    
    // 4. Create Service
    err = m.createService(ctx, namespace, vm)
    if err != nil {
        return fmt.Errorf("failed to create Service: %w", err)
    }
    
    // 5. Create Ingress (for HTTP/HTTPS)
    err = m.createIngress(ctx, namespace, vm)
    if err != nil {
        return fmt.Errorf("failed to create Ingress: %w", err)
    }
    
    return nil
}

func (m *VMManager) ensureNamespace(ctx context.Context, name string) error {
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
            Labels: map[string]string{
                "pod-security.kubernetes.io/enforce": "restricted",
                "pod-security.kubernetes.io/audit":   "restricted",
                "pod-security.kubernetes.io/warn":    "restricted",
            },
        },
    }
    
    _, err := m.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
    if errors.IsAlreadyExists(err) {
        return nil
    }
    
    return err
}

func (m *VMManager) createPVC(ctx context.Context, namespace string, vm *types.VM) error {
    storageClassName := "local-lvm"
    
    pvc := &corev1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("vm-%s-pvc", vm.ID),
            Namespace: namespace,
        },
        Spec: corev1.PersistentVolumeClaimSpec{
            AccessModes: []corev1.PersistentVolumeClaimAccessMode{
                corev1.ReadWriteOnce,
            },
            StorageClassName: &storageClassName,
            Resources: corev1.VolumeResourceRequirements{
                Requests: corev1.ResourceList{
                    corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%d", vm.Storage)),
                },
            },
        },
    }
    
    _, err := m.clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
    return err
}

func (m *VMManager) createDeployment(ctx context.Context, namespace string, vm *types.VM) error {
    trueVal := true
    falseVal := false
    
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("vm-%s", vm.ID),
            Namespace: namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: int32Ptr(1),
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{
                    "app": fmt.Sprintf("vm-%s", vm.ID),
                },
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{
                        "app": fmt.Sprintf("vm-%s", vm.ID),
                    },
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "vm",
                            Image: vm.Image,
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.2f", vm.CPU)),
                                    corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%d", vm.RAM)),
                                },
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%.2f", vm.CPU)),
                                    corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%d", vm.RAM)),
                                },
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "vm-storage",
                                    MountPath: "/",
                                },
                            },
                            SecurityContext: &corev1.SecurityContext{
                                RunAsNonRoot:             &trueVal,
                                RunAsUser:                int64Ptr(1000),
                                RunAsGroup:               int64Ptr(1000),
                                FSGroup:                  int64Ptr(1000),
                                AllowPrivilegeEscalation: &falseVal,
                                Capabilities: &corev1.Capabilities{
                                    Drop: []corev1.Capability{"ALL"},
                                },
                            },
                        },
                    },
                    Volumes: []corev1.Volume{
                        {
                            Name: "vm-storage",
                            VolumeSource: corev1.VolumeSource{
                                PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
                                    ClaimName: fmt.Sprintf("vm-%s-pvc", vm.ID),
                                },
                            },
                        },
                    },
                },
            },
        },
    }
    
    _, err := m.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
    return err
}

func (m *VMManager) createService(ctx context.Context, namespace string, vm *types.VM) error {
    service := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("vm-%s-ssh", vm.ID),
            Namespace: namespace,
        },
        Spec: corev1.ServiceSpec{
            Selector: map[string]string{
                "app": fmt.Sprintf("vm-%s", vm.ID),
            },
            Ports: []corev1.ServicePort{
                {
                    Name:       "ssh",
                    Port:       22,
                    TargetPort: intstr.FromInt(22),
                    Protocol:   corev1.ProtocolTCP,
                },
            },
            Type: corev1.ServiceTypeClusterIP,
        },
    }
    
    _, err := m.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
    return err
}

// Helper functions
func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }
func boolPtr(b bool) *bool    { return &b }
```

**Dependencies:**
```bash
# backend/go.mod
go get k8s.io/api@v0.29.0
go get k8s.io/apimachinery@v0.29.0
go get k8s.io/client-go@v0.29.0
```

---

#### Task 1.3: Backend SSH Key Generation
**Estimate:** 2 hours
**Acceptance Criteria:**
- [ ] GenerateKeyPair function creates Ed25519 keypair
- [ ] Private key in OpenSSH PEM format
- [ ] Public key in authorized_keys format
- [ ] Function tested with multiple key generations

**Implementation:**
```go
// backend/internal/ssh/keygen.go
package ssh

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/pem"
    
    "golang.org/x/crypto/ssh"
)

// GenerateKeyPair generates an Ed25519 SSH keypair
func GenerateKeyPair() (privateKey string, publicKey string, err error) {
    // Generate Ed25519 keypair
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return "", "", err
    }
    
    // Encode private key (OpenSSH PEM format)
    privateKeyPEM, err := ssh.MarshalPrivateKey(priv, "")
    if err != nil {
        return "", "", err
    }
    
    privateBytes := pem.EncodeToMemory(privateKeyPEM)
    
    // Encode public key (authorized_keys format)
    pubKey, err := ssh.NewPublicKey(pub)
    if err != nil {
        return "", "", err
    }
    
    pubBytes := ssh.MarshalAuthorizedKey(pubKey)
    
    return string(privateBytes), string(pubBytes), nil
}
```

---

#### Task 1.4: Backend Quota Enforcement
**Estimate:** 6 hours
**Acceptance Criteria:**
- [ ] CheckQuota function with SELECT FOR UPDATE
- [ ] UpdateUsage function after VM create/delete
- [ ] ReconcileUsage function (periodic DB ↔ k8s sync)
- [ ] Unit tests for race condition prevention

**Implementation:**
```go
// backend/internal/database/quota.go
package database

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    
    "github.com/podland/backend/internal/database/types"
)

var ErrQuotaExceeded = errors.New("quota exceeded")

// CheckQuota checks if user can create VM with given resources
func (db *Database) CheckQuota(ctx context.Context, userID string, cpu float64, ram, storage int64) error {
    tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelSerializable,
    })
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Lock quota row for update
    var cpuLimit, ramLimit, storageLimit float64
    var vmCountLimit int
    err = tx.QueryRowContext(ctx, `
        SELECT cpu_limit, ram_limit, storage_limit, vm_count_limit
        FROM user_quotas
        WHERE user_id = $1
        FOR UPDATE
    `, userID).Scan(&cpuLimit, &ramLimit, &storageLimit, &vmCountLimit)
    
    if err != nil {
        return fmt.Errorf("failed to get quota: %w", err)
    }
    
    // Get current usage
    var cpuUsed, ramUsed, storageUsed float64
    var vmCount int
    err = tx.QueryRowContext(ctx, `
        SELECT cpu_used, ram_used, storage_used, vm_count
        FROM user_quota_usage
        WHERE user_id = $1
    `, userID).Scan(&cpuUsed, &ramUsed, &storageUsed, &vmCount)
    
    if err != nil {
        return fmt.Errorf("failed to get usage: %w", err)
    }
    
    // Check if new VM fits
    if cpuUsed+cpu > cpuLimit {
        return ErrQuotaExceeded
    }
    if ramUsed+ram > ramLimit {
        return ErrQuotaExceeded
    }
    if storageUsed+storage > storageLimit {
        return ErrQuotaExceeded
    }
    if vmCount+1 > vmCountLimit {
        return ErrQuotaExceeded
    }
    
    return nil
}

// UpdateUsage updates quota usage after VM create/delete
func (db *Database) UpdateUsage(ctx context.Context, userID string, cpu float64, ram, storage int64, vmCountDelta int) error {
    _, err := db.db.ExecContext(ctx, `
        UPDATE user_quota_usage
        SET cpu_used = cpu_used + $1,
            ram_used = ram_used + $2,
            storage_used = storage_used + $3,
            vm_count = vm_count + $4,
            updated_at = NOW()
        WHERE user_id = $5
    `, cpu, ram, storage, vmCountDelta, userID)
    
    return err
}
```

---

### Week 2: Backend API + Frontend Core

#### Task 2.1: Backend VM Handlers
**Estimate:** 8 hours
**Acceptance Criteria:**
- [ ] POST /api/vms creates VM (returns 202 Accepted)
- [ ] GET /api/vms lists user's VMs
- [ ] GET /api/vms/{id} gets VM details
- [ ] POST /api/vms/{id}/start starts VM
- [ ] POST /api/vms/{id}/stop stops VM
- [ ] POST /api/vms/{id}/restart restarts VM
- [ ] DELETE /api/vms/{id} deletes VM
- [ ] All endpoints require JWT authentication
- [ ] All endpoints validate VM ownership

**API Endpoints:**
```go
// backend/handlers/vms.go
package handlers

type CreateVMRequest struct {
    Name string `json:"name"`
    OS   string `json:"os"`   // "ubuntu-2204" or "debian-12"
    Tier string `json:"tier"` // "nano", "micro", "small", "medium", "large", "xlarge"
}

type VMResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    OS        string    `json:"os"`
    Tier      string    `json:"tier"`
    CPU       float64   `json:"cpu"`
    RAM       int64     `json:"ram"`
    Storage   int64     `json:"storage"`
    Status    string    `json:"status"`
    Domain    string    `json:"domain"`
    CreatedAt time.Time `json:"created_at"`
}

func HandleCreateVM(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    var req CreateVMRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate tier
    tier, err := db.GetTier(req.Tier)
    if err != nil {
        http.Error(w, "Invalid tier", http.StatusBadRequest)
        return
    }
    
    // Check tier availability by role
    userRole, _ := db.GetUserRole(userID)
    if tier.MinRole == "internal" && userRole != "internal" {
        http.Error(w, "Tier not available for your role", http.StatusForbidden)
        return
    }
    
    // Check quota
    err = db.CheckQuota(r.Context(), userID, tier.CPU, tier.RAM, tier.Storage)
    if err != nil {
        if err == database.ErrQuotaExceeded {
            http.Error(w, "Quota exceeded", http.StatusForbidden)
            return
        }
        http.Error(w, "Failed to check quota", http.StatusInternalServerError)
        return
    }
    
    // Generate SSH keypair
    privateKey, publicKey, err := sshkey.GenerateKeyPair()
    if err != nil {
        http.Error(w, "Failed to generate SSH key", http.StatusInternalServerError)
        return
    }
    
    // Create VM in database
    vm := &types.VM{
        UserID:       userID,
        Name:         req.Name,
        OS:           req.OS,
        Tier:         req.Tier,
        CPU:          tier.CPU,
        RAM:          tier.RAM,
        Storage:      tier.Storage,
        Status:       "pending",
        SSHPublicKey: publicKey,
    }
    
    err = db.CreateVM(r.Context(), vm)
    if err != nil {
        http.Error(w, "Failed to create VM", http.StatusInternalServerError)
        return
    }
    
    // Update quota usage
    err = db.UpdateUsage(r.Context(), userID, tier.CPU, tier.RAM, tier.Storage, 1)
    if err != nil {
        http.Error(w, "Failed to update quota", http.StatusInternalServerError)
        return
    }
    
    // Create k8s resources (async)
    go func() {
        err := vmManager.CreateVM(context.Background(), vm)
        if err != nil {
            // Update VM status to error
            db.UpdateVMStatus(context.Background(), vm.ID, "error")
            return
        }
        
        // Update VM status to running
        db.UpdateVMStatus(context.Background(), vm.ID, "running")
    }()
    
    // Return 202 Accepted
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "id":          vm.ID,
        "status":      vm.Status,
        "ssh_key":     privateKey, // Show only once!
        "message":     "VM is being created. SSH key shown only once - download now!",
    })
}

func HandleListVMs(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    
    vms, err := db.GetUserVMs(r.Context(), userID)
    if err != nil {
        http.Error(w, "Failed to list VMs", http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(vms)
}

func HandleGetVM(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    vmID := chi.URLParam(r, "id")
    
    vm, err := db.GetVM(r.Context(), vmID, userID)
    if err != nil {
        http.Error(w, "VM not found", http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(vm)
}

func HandleStartVM(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    vmID := chi.URLParam(r, "id")
    
    vm, err := db.GetVM(r.Context(), vmID, userID)
    if err != nil {
        http.Error(w, "VM not found", http.StatusNotFound)
        return
    }
    
    if vm.Status != "stopped" {
        http.Error(w, "VM must be stopped to start", http.StatusBadRequest)
        return
    }
    
    // Scale Deployment to 1 replica
    err = vmManager.StartVM(r.Context(), vm)
    if err != nil {
        http.Error(w, "Failed to start VM", http.StatusInternalServerError)
        return
    }
    
    db.UpdateVMStatus(r.Context(), vmID, "pending")
    
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "pending",
    })
}

func HandleStopVM(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    vmID := chi.URLParam(r, "id")
    
    vm, err := db.GetVM(r.Context(), vmID, userID)
    if err != nil {
        http.Error(w, "VM not found", http.StatusNotFound)
        return
    }
    
    if vm.Status != "running" {
        http.Error(w, "VM must be running to stop", http.StatusBadRequest)
        return
    }
    
    // Scale Deployment to 0 replicas
    err = vmManager.StopVM(r.Context(), vm)
    if err != nil {
        http.Error(w, "Failed to stop VM", http.StatusInternalServerError)
        return
    }
    
    db.UpdateVMStatus(r.Context(), vmID, "stopped")
    
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "stopped",
    })
}

func HandleDeleteVM(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("user_id").(string)
    vmID := chi.URLParam(r, "id")
    
    vm, err := db.GetVM(r.Context(), vmID, userID)
    if err != nil {
        http.Error(w, "VM not found", http.StatusNotFound)
        return
    }
    
    // Create snapshot before delete (7-day retention)
    err = vmManager.CreateVMSnapshot(r.Context(), vm)
    if err != nil {
        log.Printf("Failed to create snapshot: %v", err)
        // Continue with delete anyway
    }
    
    // Delete k8s resources
    err = vmManager.DeleteVM(r.Context(), vm)
    if err != nil {
        http.Error(w, "Failed to delete VM", http.StatusInternalServerError)
        return
    }
    
    // Update quota usage
    err = db.UpdateUsage(r.Context(), userID, -vm.CPU, -vm.RAM, -vm.Storage, -1)
    if err != nil {
        http.Error(w, "Failed to update quota", http.StatusInternalServerError)
        return
    }
    
    // Delete from database
    err = db.DeleteVM(r.Context(), vmID)
    if err != nil {
        http.Error(w, "Failed to delete VM", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

---

### Week 3: Frontend Implementation

#### Task 3.1: Frontend VM List View
**Estimate:** 4 hours

#### Task 3.2: Frontend VM Detail Page
**Estimate:** 6 hours

#### Task 3.3: Frontend Create VM Wizard
**Estimate:** 6 hours

---

### Week 4: Testing + Documentation

#### Task 4.1: Integration Testing
**Estimate:** 4 hours

#### Task 4.2: Load Testing
**Estimate:** 2 hours

#### Task 4.3: Documentation
**Estimate:** 2 hours

---

## Files Summary

| File | Purpose |
|------|---------|
| `backend/internal/k8s/vm_manager.go` | k3s client for VM lifecycle |
| `backend/internal/ssh/keygen.go` | SSH keypair generation |
| `backend/internal/database/quota.go` | Quota enforcement |
| `backend/internal/database/vms.go` | VM database operations |
| `backend/handlers/vms.go` | VM API endpoints |
| `backend/migrations/002_phase2_vm_quota.sql` | Database schema |
| `frontend/src/routes/dashboard/vms.tsx` | VM list page |
| `frontend/src/routes/dashboard/vms/$id.tsx` | VM detail page |
| `frontend/src/components/vm/CreateVMWizard.tsx` | Create VM form |
| `infra/k3s/traefik-config.yaml` | Traefik configuration |
| `.github/workflows/build-vm-images.yml` | VM image builds |

---

## Acceptance Criteria

- [ ] All 16 success criteria met
- [ ] All 12 technical milestones completed
- [ ] All API endpoints tested
- [ ] All frontend components tested
- [ ] Load test passed (100 concurrent VM creations)
- [ ] Documentation complete

---

*Plan created: 2026-03-27*
*Ready for implementation*
