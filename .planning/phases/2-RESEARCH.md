# Phase 2 Research: Core VM

**Phase:** 2 of 5
**Goal:** Users can create and manage VMs with resource quotas enforced
**Research Date:** 2026-03-27
**Status:** Complete

---

## Research Questions (from CONTEXT.md)

### 1. k3s client-go Patterns ✅

**Question:** What's the Go library pattern for creating Deployments, PVCs, Services programmatically?

**Answer:** Use Kubernetes `client-go` library with dynamic client for resource creation.

**Implementation Pattern:**
```go
// backend/internal/k8s/vm_manager.go
package k8s

import (
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    networkingv1 "k8s.io/api/networking/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

type VMManager struct {
    clientset *kubernetes.Clientset
}

func NewVMManager(kubeconfig string) (*VMManager, error) {
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, err
    }
    
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }
    
    return &VMManager{clientset: clientset}, nil
}

// CreateVM creates all k8s resources for a VM
func (m *VMManager) CreateVM(ctx context.Context, vm *types.VM) error {
    // 1. Create namespace (if not exists)
    err := m.ensureNamespace(ctx, vm.UserID)
    if err != nil {
        return err
    }
    
    // 2. Create PVC
    err = m.createPVC(ctx, vm)
    if err != nil {
        return err
    }
    
    // 3. Create Deployment
    err = m.createDeployment(ctx, vm)
    if err != nil {
        return err
    }
    
    // 4. Create Service
    err = m.createService(ctx, vm)
    if err != nil {
        return err
    }
    
    // 5. Create Ingress
    err = m.createIngress(ctx, vm)
    if err != nil {
        return err
    }
    
    return nil
}

func (m *VMManager) createPVC(ctx context.Context, vm *types.VM) error {
    pvc := &corev1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("vm-%s-pvc", vm.ID),
            Namespace: fmt.Sprintf("user-%s", vm.UserID),
        },
        Spec: corev1.PersistentVolumeClaimSpec{
            AccessModes: []corev1.PersistentVolumeClaimAccessMode{
                corev1.ReadWriteOnce,
            },
            StorageClassName: &vm.StorageClass, // "local-lvm"
            Resources: corev1.VolumeResourceRequirements{
                Requests: corev1.ResourceList{
                    corev1.ResourceStorage: resource.MustParse(vm.Storage),
                },
            },
        },
    }
    
    _, err := m.clientset.CoreV1().PersistentVolumeClaims(
        fmt.Sprintf("user-%s", vm.UserID),
    ).Create(ctx, pvc, metav1.CreateOptions{})
    
    return err
}

func (m *VMManager) createDeployment(ctx context.Context, vm *types.VM) error {
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      fmt.Sprintf("vm-%s", vm.ID),
            Namespace: fmt.Sprintf("user-%s", vm.UserID),
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
                            Image: vm.Image, // e.g., "podland/ubuntu-2204:v1.0"
                            Resources: corev1.ResourceRequirements{
                                Limits: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(vm.CPU),
                                    corev1.ResourceMemory: resource.MustParse(vm.RAM),
                                },
                                Requests: corev1.ResourceList{
                                    corev1.ResourceCPU:    resource.MustParse(vm.CPU),
                                    corev1.ResourceMemory: resource.MustParse(vm.RAM),
                                },
                            },
                            VolumeMounts: []corev1.VolumeMount{
                                {
                                    Name:      "vm-storage",
                                    MountPath: "/",
                                },
                            },
                            SecurityContext: &corev1.SecurityContext{
                                RunAsNonRoot:             boolPtr(true),
                                RunAsUser:                int64Ptr(1000),
                                RunAsGroup:               int64Ptr(1000),
                                AllowPrivilegeEscalation: boolPtr(false),
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
    
    _, err := m.clientset.AppsV1().Deployments(
        fmt.Sprintf("user-%s", vm.UserID),
    ).Create(ctx, deployment, metav1.CreateOptions{})
    
    return err
}
```

**Dependencies:**
```go
// go.mod
require (
    k8s.io/api v0.29.0
    k8s.io/apimachinery v0.29.0
    k8s.io/client-go v0.29.0
)
```

**References:**
- https://github.com/kubernetes/client-go
- https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/

---

### 2. Pre-built OS Images ✅

**Question:** What's the best way to build and host Ubuntu 22.04 and Debian 12 container images?

**Answer:** Use GitHub Actions to build from official images, host on Docker Hub.

**Dockerfile (Ubuntu 22.04):**
```dockerfile
# vm-images/ubuntu-2204/Dockerfile
FROM ubuntu:22.04

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Install essential tools
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    vim \
    htop \
    net-tools \
    iputils-ping \
    dnsutils \
    ca-certificates \
    gnupg \
    lsb-release \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user (UID 1000)
RUN useradd -m -u 1000 -s /bin/bash vmuser

# Set working directory
WORKDIR /home/vmuser

# Switch to non-root user
USER vmuser

# Default command
CMD ["/bin/bash"]
```

**GitHub Actions Workflow:**
```yaml
# .github/workflows/build-vm-images.yml
name: Build VM Images

on:
  schedule:
    - cron: '0 0 1 * *'  # Monthly (1st of every month)
  workflow_dispatch:     # Manual trigger

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - os: ubuntu-2204
            context: ./vm-images/ubuntu-2204
            tags: podland/ubuntu-2204
          - os: debian-12
            context: ./vm-images/debian-12
            tags: podland/debian-12

    steps:
      - uses: actions/checkout@v4

      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push ${{ matrix.os }}
        uses: docker/build-push-action@v5
        with:
          context: ${{ matrix.context }}
          push: true
          tags: |
            ${{ matrix.tags }}:latest
            ${{ matrix.tags }}:v${{ github.run_number }}
            ${{ matrix.tags }}:${{ github.run_started_at.strftime('%Y-%m') }}
```

**Image Size Estimates:**
- ubuntu:22.04 (official): 77MB
- With essentials: ~300MB
- debian:12 (official): 124MB
- With essentials: ~320MB

**References:**
- https://hub.docker.com/_/ubuntu
- https://hub.docker.com/_/debian
- https://docs.github.com/en/actions/publishing-packages/publishing-docker-images

---

### 3. Ingress Controller (Traefik) ✅

**Question:** k3s comes with Traefik, but do we need to configure it specially for dynamic subdomain routing?

**Answer:** Yes, need to configure Traefik for wildcard subdomain routing and TCP passthrough (SSH).

**Traefik Configuration for k3s:**
```yaml
# infra/k3s/traefik-config.yaml
apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: traefik
  namespace: kube-system
spec:
  valuesContent: |
    ports:
      web:
        exposedPort: 80
        protocol: HTTP
      websecure:
        exposedPort: 443
        protocol: HTTPS
      ssh:
        exposedPort: 22
        protocol: TCP
      game:
        exposedPort: 25565
        protocol: TCP
    providers:
      kubernetesCRD:
        enabled: true
      kubernetesIngress:
        enabled: true
```

**IngressRouteTCP for SSH:**
```yaml
# infra/k3s/vm-ssh-ingress.yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: vm-ssh-ingress
  namespace: podland
spec:
  entryPoints:
    - ssh
  routes:
    - match: HostSNI(`*.podland.app`)
      services:
        - name: vm-ssh-router
          port: 22
```

**Wildcard Ingress for HTTP/HTTPS:**
```yaml
# infra/k3s/vm-http-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: vm-http-ingress
  namespace: podland
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: web,websecure
    traefik.ingress.kubernetes.io/router.tls: "true"
spec:
  rules:
  - host: "*.podland.app"
    http:
      paths:
      - pathType: Prefix
        path: /
        backend:
          service:
            name: vm-router
            port:
              number: 80
```

**Note:** k3s Traefik v2 supports wildcard hosts via IngressRouteTCP. For HTTP/HTTPS, may need to use specific subdomain per VM (vm-name.podland.app).

**References:**
- https://doc.traefik.io/traefik/providers/kubernetes-crd/
- https://github.com/traefik/traefik-k8s-examples
- https://k3s.io/docs/networking/ingress/

---

### 4. ResourceQuota + PVC Storage ✅

**Question:** How does k3s ResourceQuota interact with PVC storage limits? Any gotchas?

**Answer:** ResourceQuota can limit PVC count and total storage, but requires careful configuration.

**ResourceQuota Configuration:**
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: user-quota
  namespace: user-{user-id}
spec:
  hard:
    # CPU limits
    requests.cpu: "4"
    limits.cpu: "4"
    
    # Memory limits
    requests.memory: 8Gi
    limits.memory: 8Gi
    
    # Storage limits
    persistentvolumeclaims: "5"  # Max number of PVCs
    requests.storage: 100Gi      # Total storage across all PVCs
    
    # Service limits
    services.loadbalancers: "1"
    services.nodeports: "10"
```

**Gotchas:**

1. **PVC Storage Accounting:**
   - `requests.storage` counts total storage across ALL PVCs in namespace
   - If user creates 2 VMs (5GB + 10GB), total = 15GB counted
   - Quota check happens BEFORE PVC creation

2. **StorageClass Specific Limits:**
   ```yaml
   # Can limit per storage class
   requests.storage: 100Gi              # Total across all classes
   local-lvm.requests.storage: 100Gi    # Specific to local-lvm
   ```

3. **PVC Deletion Lag:**
   - When PVC deleted, quota not immediately freed
   - PVC stays in "Terminating" until underlying storage released
   - Can cause "quota exceeded" errors temporarily

4. **Best Practice:**
   - Track quota in database (source of truth)
   - Use k8s ResourceQuota as secondary enforcement
   - Reconcile DB ↔ k8s periodically (every 5 min)

**References:**
- https://kubernetes.io/docs/concepts/policy/resource-quotas/
- https://kubernetes.io/docs/concepts/storage/persistent-volumes/
- https://k3s.io/docs/admission/

---

### 5. PodSecurityStandard (Restricted) ✅

**Question:** What specific restrictions does `restricted` level enforce? Will our VM containers pass validation?

**Answer:** `restricted` level enforces strict security requirements. Our VM containers will pass with proper configuration.

**PodSecurityStandard Levels:**

| Level | Description | Use Case |
|-------|-------------|----------|
| `privileged` | No restrictions | System pods, kube-system |
| `baseline` | Minimally restrictive | Most workloads |
| `restricted` | Heavily restricted | Multi-tenant, security-sensitive |

**Restricted Level Requirements:**

1. **Volume Types:** Only safe volume types allowed
   - ✅ persistentVolumeClaim, emptyDir, configMap, secret, projected
   - ❌ hostPath, hostIPC, hostNetwork, hostPID

2. **Privileged Containers:**
   - ❌ `privileged: true` not allowed
   - ✅ `privileged: false` (default)

3. **Capabilities:**
   - ❌ Cannot add capabilities
   - ✅ Must drop ALL capabilities

4. **Privilege Escalation:**
   - ❌ `allowPrivilegeEscalation: true` not allowed
   - ✅ `allowPrivilegeEscalation: false` required

5. **Run As Non-Root:**
   - ❌ `runAsNonRoot: false` not allowed
   - ✅ `runAsNonRoot: true` required

6. **Seccomp Profile:**
   - ✅ Must set seccomp profile to `RuntimeDefault` or specific profile

**VM Container Configuration (Passes Validation):**
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: false  # Users need to write files
  seccompProfile:
    type: RuntimeDefault
  capabilities:
    drop:
      - ALL
```

**Namespace Label for Enforcement:**
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: user-{user-id}
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Testing:**
```bash
# Test if pod spec passes restricted PSS
kubectl apply -f vm-deployment.yaml --dry-run=server

# Check audit logs for violations
kubectl logs -n kube-system -l app=traefik | grep -i security
```

**References:**
- https://kubernetes.io/docs/concepts/security/pod-security-standards/
- https://kubernetes.io/docs/concepts/security/pod-security-admission/

---

## Additional Research Findings

### 6. SSH Key Generation (Go)

**Library:** `golang.org/x/crypto/ssh`

```go
package ssh

import (
    "crypto/rand"
    "crypto/ed25519"
    "encoding/pem"
    "golang.org/x/crypto/ssh"
)

func GenerateKeyPair() (privateKey string, publicKey string, err error) {
    // Generate Ed25519 keypair
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        return "", "", err
    }
    
    // Encode private key (OpenSSH format)
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

**References:**
- https://pkg.go.dev/golang.org/x/crypto/ssh
- https://coolaj86.com/articles/generate-and-open-ssh-keys-in-golang/

---

### 7. Cloud-Init for VM Configuration

**Question:** How to inject SSH public key into VM on first boot?

**Answer:** Use cloud-init with user-data.

**Cloud-Init Configuration:**
```yaml
#cloud-config
users:
  - name: vmuser
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - {{.SSHPublicKey}}

packages:
  - git
  - curl
  - wget

runcmd:
  - echo "Welcome to Podland VM!"
  - touch /home/vmuser/.podland-initialized

final_message: |
  The system is finally up, after $UPTIME seconds
```

**Implementation in k8s:**
```yaml
# Option 1: ConfigMap + Volume Mount
apiVersion: v1
kind: ConfigMap
metadata:
  name: vm-{vm-id}-cloud-init
  namespace: user-{user-id}
data:
  user-data: |
    #cloud-config
    users:
      - name: vmuser
        ssh_authorized_keys:
          - {{.SSHPublicKey}}
---
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: vm
    volumeMounts:
    - name: cloud-init
      mountPath: /var/lib/cloud/seed/nocloud
  volumes:
  - name: cloud-init
    configMap:
      name: vm-{vm-id}-cloud-init
```

**Note:** cloud-init works best with cloud images. For simple containers, can use entrypoint script instead.

**References:**
- https://cloudinit.readthedocs.io/
- https://cloudinit.readthedocs.io/en/latest/reference/datasources/nocloud.html

---

## Summary: Research Findings

| Question | Answer | Confidence |
|----------|--------|------------|
| k3s client-go patterns | Use dynamic client, create Deployment/PVC/Service/Ingress | ✅ High |
| Pre-built OS images | GitHub Actions + Docker Hub (ubuntu:22.04 + essentials) | ✅ High |
| Ingress controller | Traefik (built-in k3s), configure for wildcard + TCP | ✅ High |
| ResourceQuota + PVC | Can limit PVC count + total storage, track in DB | ✅ High |
| PodSecurityStandard | `restricted` level, VM containers will pass with proper config | ✅ High |
| SSH key generation | Ed25519 via `golang.org/x/crypto/ssh` | ✅ High |
| VM configuration | cloud-init or entrypoint script for SSH key injection | ✅ Medium |

---

## Recommended Next Steps

1. **Create 2-PLAN.md** — Implementation plan with tasks
2. **Database schema migration** — Add vms, user_quotas, tiers tables
3. **Backend k8s module** — Implement VMManager with client-go
4. **Frontend VM wizard** — Create VM UI with tier selection
5. **GitHub Actions workflow** — Build VM images monthly

---

*Research conducted: 2026-03-27*
*Confidence: High — All questions answered with implementation patterns*
