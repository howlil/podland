# Phase 5 Plan: Admin + Polish

**Phase:** 5 of 5
**Goal:** Superadmin can manage platform, API is complete, ready for launch
**Requirements:** 10 (ADMIN-01 through ADMIN-05, IDLE-01 through IDLE-04, VM-08)
**Duration:** 4 weeks
**Status:** Ready for execution

---

## Week Breakdown

| Week | Focus | Tasks | Estimated Hours |
|------|-------|-------|-----------------|
| 1 | Admin Panel Backend | 1.1-1.4 | 18h |
| 2 | Idle Detection + Email | 2.1-2.4 | 20h |
| 3 | Frontend Admin + Pin UI | 3.1-3.4 | 16h |
| 4 | Load Testing + Backup | 4.1-4.4 | 16h |

**Total:** 70 hours (~2 weeks full-time or 4 weeks part-time)

---

## Week 1: Admin Panel Backend

### Task 1.1: Admin Authorization Middleware (4h)

**Objective:** Create middleware to restrict admin routes to superadmin users only.

**Files to Create:**
```
apps/backend/internal/middleware/admin.go
```

**Implementation:**
```go
// internal/middleware/admin.go
package middleware

import (
    "context"
    "net/http"
    
    "github.com/podland/backend/internal/repository"
)

// AdminOnly restricts access to superadmin users only
func AdminOnly(userRepo repository.UserRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID, ok := r.Context().Value("user_id").(string)
            if !ok || userID == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            
            user, err := userRepo.GetByID(r.Context(), userID)
            if err != nil || user.Role != "superadmin" {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            // Add user to context for downstream handlers
            ctx := context.WithValue(r.Context(), "user", user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**Acceptance Criteria:**
- [ ] Middleware returns 401 for unauthenticated requests
- [ ] Middleware returns 403 for non-superadmin users
- [ ] Superadmin users can access protected routes
- [ ] User object added to context for handlers

**Dependencies:** None (uses existing User entity and repository)

---

### Task 1.2: Audit Logging Middleware (4h)

**Objective:** Create middleware to automatically log all admin actions.

**Files to Create:**
```
apps/backend/internal/middleware/audit.go
apps/backend/internal/repository/audit_repository.go
apps/backend/internal/entity/audit_log.go
```

**Implementation:**
```go
// internal/middleware/audit.go
func AuditLogger(auditRepo repository.AuditRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID, _ := r.Context().Value("user_id").(string)
            action := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
            ip := r.RemoteAddr
            userAgent := r.UserAgent()
            
            // Log asynchronously (don't block request)
            go auditRepo.Create(r.Context(), &entity.AuditLog{
                UserID:    userID,
                Action:    action,
                IPAddress: ip,
                UserAgent: userAgent,
            })
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Database Migration:**
```sql
-- migrations/005_phase5_admin.sql
CREATE TABLE audit_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id),
  action VARCHAR(255) NOT NULL,
  ip_address INET,
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

**Acceptance Criteria:**
- [ ] Every admin request creates audit log entry
- [ ] Logging is asynchronous (doesn't block request)
- [ ] Audit logs include user ID, action, IP, user agent
- [ ] Errors in logging don't break admin requests

**Dependencies:** Task 1.1

---

### Task 1.3: Admin Handlers (6h)

**Objective:** Create handlers for admin panel functionality.

**Files to Create:**
```
apps/backend/internal/handler/admin_handler.go
```

**Implementation:**
```go
// internal/handler/admin_handler.go
type AdminHandler struct {
    userRepo     repository.UserRepository
    auditRepo    repository.AuditRepository
    vmRepo       repository.VMRepository
}

// ListUsers returns all users with optional role filter
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    role := r.URL.Query().Get("role") // optional filter
    
    var users []entity.User
    var err error
    if role != "" {
        users, err = h.userRepo.GetByRole(r.Context(), role)
    } else {
        users, err = h.userRepo.GetAll(r.Context())
    }
    
    pkgresponse.JSON(w, http.StatusOK, users)
}

// ChangeRole changes a user's role (superadmin only)
func (h *AdminHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    var req struct {
        Role string `json:"role"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        pkgresponse.BadRequest(w, "Invalid request body")
        return
    }
    
    // Validate role
    validRoles := map[string]bool{
        "internal": true,
        "external": true,
        "superadmin": true,
    }
    
    if !validRoles[req.Role] {
        pkgresponse.BadRequest(w, "Invalid role")
        return
    }
    
    err := h.userRepo.UpdateRole(r.Context(), userID, req.Role)
    if err != nil {
        pkgresponse.Error(w, err, "Failed to update role")
        return
    }
    
    pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// BanUser bans a user
func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    
    err := h.userRepo.BanUser(r.Context(), userID)
    if err != nil {
        pkgresponse.Error(w, err, "Failed to ban user")
        return
    }
    
    pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// SystemHealth returns cluster health metrics
func (h *AdminHandler) SystemHealth(w http.ResponseWriter, r *http.Request) {
    // Query Prometheus for cluster metrics
    health := map[string]interface{}{
        "cluster_cpu":    45.2,  // %
        "cluster_memory": 62.5,  // %
        "cluster_storage": 38.1, // %
        "total_users":    487,
        "total_vms":      234,
        "active_vms":     156,
    }
    
    pkgresponse.JSON(w, http.StatusOK, health)
}

// AuditLog returns audit log entries
func (h *AdminHandler) AuditLog(w http.ResponseWriter, r *http.Request) {
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit == 0 {
        limit = 100
    }
    
    logs, err := h.auditRepo.GetRecent(r.Context(), limit)
    if err != nil {
        pkgresponse.Error(w, err, "Failed to fetch audit logs")
        return
    }
    
    pkgresponse.JSON(w, http.StatusOK, logs)
}
```

**Acceptance Criteria:**
- [ ] GET /api/admin/users returns all users
- [ ] GET /api/admin/users?role=internal returns filtered users
- [ ] PATCH /api/admin/users/{id}/role changes role successfully
- [ ] POST /api/admin/users/{id}/ban bans user
- [ ] GET /api/admin/health returns cluster metrics
- [ ] GET /api/admin/audit-log returns recent audit logs

**Dependencies:** Tasks 1.1, 1.2

---

### Task 1.4: Wire Up Admin Routes (4h)

**Objective:** Add admin routes to main router with proper middleware.

**Files to Modify:**
```
apps/backend/cmd/main.go
```

**Implementation:**
```go
// cmd/main.go
func main() {
    // ... existing setup ...
    
    // Create repositories
    userRepo := repository.NewUserRepository(db)
    auditRepo := repository.NewAuditRepository(db)
    
    // Create handlers
    adminHandler := handler.NewAdminHandler(userRepo, auditRepo, vmRepo)
    
    // Admin routes (protected)
    r.Route("/api/admin", func(r chi.Router) {
        r.Use(middleware.Auth(authHelper))        // Require authentication
        r.Use(middleware.AdminOnly(userRepo))     // Require superadmin role
        r.Use(middleware.AuditLogger(auditRepo))  // Auto-log all actions
        
        r.Get("/users", adminHandler.ListUsers)
        r.Patch("/users/{id}/role", adminHandler.ChangeRole)
        r.Post("/users/{id}/ban", adminHandler.BanUser)
        r.Get("/health", adminHandler.SystemHealth)
        r.Get("/audit-log", adminHandler.AuditLog)
    })
    
    // ... rest of setup ...
}
```

**Acceptance Criteria:**
- [ ] Admin routes accessible only to superadmin
- [ ] All admin actions logged to audit_logs table
- [ ] Routes return 401/403 for unauthorized users
- [ ] Swagger/docs updated with admin endpoints

**Dependencies:** Tasks 1.1, 1.2, 1.3

---

## Week 2: Idle Detection + Email

### Task 2.1: Idle Detector Service (6h)

**Objective:** Create service to detect idle VMs using Prometheus metrics.

**Files to Create:**
```
apps/backend/internal/idle/detector.go
apps/backend/internal/idle/service.go
```

**Implementation:**
```go
// internal/idle/detector.go
package idle

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/podland/backend/internal/entity"
    "github.com/podland/backend/internal/repository"
)

type Detector struct {
    prometheusURL  string
    vmRepo         repository.VMRepository
    notificationRepo repository.NotificationRepository
    httpClient     *http.Client
}

type PrometheusResponse struct {
    Data struct {
        Result []struct {
            Metric struct {
                VMID string `json:"vm_id"`
            } `json:"metric"`
            Value []interface{} `json:"value"`
        } `json:"result"`
    } `json:"data"`
}

func NewDetector(prometheusURL string, vmRepo repository.VMRepository, notificationRepo repository.NotificationRepository) *Detector {
    return &Detector{
        prometheusURL:  prometheusURL,
        vmRepo:         vmRepo,
        notificationRepo: notificationRepo,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// Run executes idle detection
func (d *Detector) Run() {
    ctx := context.Background()
    
    // Query 1: No HTTP traffic for 48h
    idleHTTP, _ := d.queryPrometheus(`sum_over_time(container_network_receive_bytes_total[48h]) == 0`)
    
    // Query 2: No CPU activity for 48h
    idleCPU, _ := d.queryPrometheus(`avg_over_time(container_cpu_usage_seconds_total[48h]) < 0.01`)
    
    // Find intersection (VMs idle on both metrics)
    idleVMIDs := d.findIntersection(idleHTTP, idleCPU)
    
    for _, vmID := range idleVMIDs {
        vm, err := d.vmRepo.GetByID(ctx, vmID)
        if err != nil || vm == nil {
            continue
        }
        
        // Skip pinned VMs
        if vm.IsPinned {
            continue
        }
        
        // Check if already warned
        if vm.IdleWarnedAt != nil && time.Since(*vm.IdleWarnedAt) > 24*time.Hour {
            // Delete VM
            d.deleteVM(vm)
        } else if vm.IdleWarnedAt == nil {
            // Send warning
            d.sendWarning(vm)
        }
    }
}

func (d *Detector) queryPrometheus(query string) ([]string, error) {
    url := fmt.Sprintf("%s/api/v1/query?query=%s", d.prometheusURL, query)
    
    resp, err := d.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    
    var promResp PrometheusResponse
    if err := json.Unmarshal(body, &promResp); err != nil {
        return nil, err
    }
    
    var vmIDs []string
    for _, result := range promResp.Data.Result {
        vmIDs = append(vmIDs, result.Metric.VMID)
    }
    
    return vmIDs, nil
}

func (d *Detector) findIntersection(lists ...[]string) []string {
    if len(lists) == 0 {
        return nil
    }
    
    countMap := make(map[string]int)
    for _, list := range lists {
        seen := make(map[string]bool)
        for _, id := range list {
            if !seen[id] {
                countMap[id]++
                seen[id] = true
            }
        }
    }
    
    var result []string
    for id, count := range countMap {
        if count == len(lists) {
            result = append(result, id)
        }
    }
    
    return result
}
```

**Acceptance Criteria:**
- [ ] Detector queries Prometheus for idle VMs
- [ ] Combined criteria (HTTP + CPU) used
- [ ] Pinned VMs skipped
- [ ] Warning sent for first-time idle VMs
- [ ] VM deleted after 24h warning

**Dependencies:** None (uses existing Prometheus from Phase 4)

---

### Task 2.2: Email Service Integration (6h)

**Objective:** Integrate SendGrid for email notifications.

**Files to Create:**
```
apps/backend/internal/email/service.go
apps/backend/internal/email/templates/idle_warning.html
apps/backend/internal/email/templates/idle_warning.txt
```

**Implementation:**
```go
// internal/email/service.go
package email

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
    client    *sendgrid.Client
    fromEmail string
}

func NewEmailService() *EmailService {
    apiKey := os.Getenv("SENDGRID_API_KEY")
    fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
    
    return &EmailService{
        client:    sendgrid.NewSendClient(apiKey),
        fromEmail: fromEmail,
    }
}

func (s *EmailService) SendIdleWarning(userEmail, userName, vmName, vmID string) error {
    deleteAt := time.Now().Add(24 * time.Hour)
    
    // Render templates
    htmlBody := renderHTMLTemplate("idle_warning.html", map[string]interface{}{
        "UserName": userName,
        "VMName":   vmName,
        "VMID":     vmID,
        "DeleteAt": deleteAt.Format("2006-01-02 15:04:05"),
    })
    
    textBody := renderTextTemplate("idle_warning.txt", map[string]interface{}{
        "UserName": userName,
        "VMName":   vmName,
        "VMID":     vmID,
        "DeleteAt": deleteAt.Format("2006-01-02 15:04:05"),
    })
    
    // Create email
    from := mail.NewEmail("Podland", s.fromEmail)
    to := mail.NewEmail(userName, userEmail)
    subject := "Your VM will be deleted in 24 hours"
    
    message := mail.NewV3Mail()
    message.SetFrom(from)
    message.AddMailPersonalization(mail.NewPersonalization())
    message.Personalizations[0].AddTos(to)
    message.SetSubject(subject)
    message.AddContent(mail.NewContent("text/html", htmlBody))
    message.AddContent(mail.NewContent("text/plain", textBody))
    
    // Send with retry
    return s.sendWithRetry(message, 3)
}

func (s *EmailService) sendWithRetry(message *mail.V3Mail, maxRetries int) error {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        response, err := s.client.Send(message)
        if err == nil && response.StatusCode == 202 {
            return nil
        }
        
        lastErr = err
        log.Printf("Email send failed (attempt %d/%d): %v", i+1, maxRetries, err)
        
        // Exponential backoff
        backoff := time.Duration(i*i) * time.Minute
        time.Sleep(backoff)
    }
    
    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

**Environment Variables:**
```bash
# Email configuration
SENDGRID_API_KEY=SG.xxxxx
SENDGRID_FROM_EMAIL=noreply@podland.app
```

**Acceptance Criteria:**
- [ ] SendGrid client initialized with API key
- [ ] Multipart email (HTML + text) sent successfully
- [ ] Retry logic with exponential backoff (3 retries)
- [ ] Email templates render correctly with variables

**Dependencies:** None (external SendGrid API)

---

### Task 2.3: Idle Detection Cron Job (4h)

**Objective:** Schedule idle detection to run hourly.

**Files to Modify:**
```
apps/backend/cmd/main.go
```

**Implementation:**
```go
// cmd/main.go
func main() {
    // ... existing setup ...
    
    // Create idle detector
    prometheusURL := os.Getenv("PROMETHEUS_URL")
    detector := idle.NewDetector(prometheusURL, vmRepo, notificationRepo)
    
    // Schedule hourly idle detection
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        for range ticker.C {
            log.Println("Running idle VM detection...")
            detector.Run()
        }
    }()
    
    // ... rest of setup ...
}
```

**Acceptance Criteria:**
- [ ] Idle detection runs every hour
- [ ] Logs execution start time
- [ ] Errors in detection don't crash server
- [ ] Multiple runs don't cause goroutine leaks

**Dependencies:** Task 2.1

---

### Task 2.4: Pin VM Feature (4h)

**Objective:** Allow users to pin VMs to prevent auto-deletion.

**Files to Create/Modify:**
```
apps/backend/internal/handler/vm_handler.go (modify)
apps/backend/internal/repository/vm_repository.go (modify)
apps/backend/migrations/005_phase5_admin.sql (modify)
```

**Database Migration:**
```sql
-- Add to migrations/005_phase5_admin.sql
ALTER TABLE vms ADD COLUMN is_pinned BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_vms_is_pinned ON vms(is_pinned);
```

**Implementation:**
```go
// internal/handler/vm_handler.go
// Add new handlers
func (h *vmHandler) PinVM(w http.ResponseWriter, r *http.Request) {
    userID := h.authHelper.GetAuthUserID(r)
    vmID := chi.URLParam(r, "id")
    
    user, _ := h.userRepo.GetByID(r.Context(), userID)
    
    // Check pin limit
    limit := 1
    if user.Role == "internal" {
        limit = 3
    }
    
    pinnedCount := h.vmRepo.GetPinnedCount(r.Context(), userID)
    if pinnedCount >= limit {
        pkgresponse.BadRequest(w, fmt.Sprintf("Pin limit exceeded (max %d)", limit))
        return
    }
    
    err := h.vmRepo.SetPinned(r.Context(), vmID, true)
    if err != nil {
        pkgresponse.Error(w, err, "Failed to pin VM")
        return
    }
    
    pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *vmHandler) UnpinVM(w http.ResponseWriter, r *http.Request) {
    userID := h.authHelper.GetAuthUserID(r)
    vmID := chi.URLParam(r, "id")
    
    err := h.vmRepo.SetPinned(r.Context(), vmID, false)
    if err != nil {
        pkgresponse.Error(w, err, "Failed to unpin VM")
        return
    }
    
    pkgresponse.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
```

**Acceptance Criteria:**
- [ ] POST /api/vms/{id}/pin pins VM
- [ ] DELETE /api/vms/{id}/pin unpins VM
- [ ] Pin limit enforced (External: 1, Internal: 3)
- [ ] Pinned VMs excluded from idle detection

**Dependencies:** Task 2.1 (idle detector checks is_pinned)

---

## Week 3: Frontend Admin + Pin UI

### Task 3.1: Admin Dashboard Page (4h)

**Objective:** Create admin dashboard with navigation to all admin features.

**Files to Create:**
```
apps/frontend/src/routes/admin/index.tsx
```

**Implementation:**
```tsx
// routes/admin/index.tsx
import { createFileRoute } from '@tanstack/react-router'
import { DashboardLayout } from '~/components/layout/DashboardLayout'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'

export const Route = createFileRoute('/admin/')({
  component: AdminDashboard,
})

function AdminDashboard() {
  return (
    <DashboardLayout>
      <div className="container mx-auto p-6">
        <h1 className="text-3xl font-bold mb-6">Admin Panel</h1>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>User Management</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-600 mb-4">
                View and manage all users, change roles, ban/unban
              </p>
              <Button asChild>
                <Link to="/admin/users">Manage Users →</Link>
              </Button>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>System Health</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-600 mb-4">
                Monitor cluster CPU, memory, and storage usage
              </p>
              <Button asChild variant="outline">
                <Link to="/admin/health">View Health →</Link>
              </Button>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Audit Log</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-gray-600 mb-4">
                View all admin actions with timestamps
              </p>
              <Button asChild variant="outline">
                <Link to="/admin/audit-log">View Logs →</Link>
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </DashboardLayout>
  )
}
```

**Acceptance Criteria:**
- [ ] Admin dashboard accessible at /admin
- [ ] Three cards with links to sub-pages
- [ ] Responsive layout (mobile → desktop)
- [ ] Only visible to superadmin users

**Dependencies:** None (static page initially)

---

### Task 3.2: User Management Page (6h)

**Objective:** Create user management interface with role change and ban functionality.

**Files to Create:**
```
apps/frontend/src/routes/admin/users.tsx
apps/frontend/src/components/admin/UserTable.tsx
```

**Implementation:**
```tsx
// routes/admin/users.tsx
import { createFileRoute, useSearch } from '@tanstack/react-router'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '~/lib/api'
import { DashboardLayout } from '~/components/layout/DashboardLayout'
import { UserTable } from '~/components/admin/UserTable'

export const Route = createFileRoute('/admin/users')({
  component: AdminUsersPage,
})

function AdminUsersPage() {
  const search = useSearch({ from: '/admin/users' })
  const [roleFilter, setRoleFilter] = useState(search.role || '')
  const queryClient = useQueryClient()
  
  const { data: users, isLoading } = useQuery({
    queryKey: ['admin-users', roleFilter],
    queryFn: () => api.get(`/api/admin/users${roleFilter ? `?role=${roleFilter}` : ''}`),
  })
  
  const changeRoleMutation = useMutation({
    mutationFn: ({ userID, role }: { userID: string; role: string }) =>
      api.patch(`/api/admin/users/${userID}/role`, { role }),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-users'])
    },
  })
  
  const banUserMutation = useMutation({
    mutationFn: (userID: string) =>
      api.post(`/api/admin/users/${userID}/ban`),
    onSuccess: () => {
      queryClient.invalidateQueries(['admin-users'])
    },
  })
  
  return (
    <DashboardLayout>
      <div className="container mx-auto p-6">
        <div className="flex justify-between items-center mb-6">
          <h1 className="text-3xl font-bold">User Management</h1>
          
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value)}
            className="border rounded px-3 py-2"
          >
            <option value="">All Roles</option>
            <option value="internal">Internal</option>
            <option value="external">External</option>
            <option value="superadmin">Superadmin</option>
          </select>
        </div>
        
        {isLoading ? (
          <div>Loading...</div>
        ) : (
          <UserTable
            users={users}
            onChangeRole={changeRoleMutation.mutate}
            onBanUser={banUserMutation.mutate}
          />
        )}
      </div>
    </DashboardLayout>
  )
}
```

**Acceptance Criteria:**
- [ ] User table shows all users with email, role, NIM
- [ ] Role filter dropdown (All/Internal/External/Superadmin)
- [ ] Role change dropdown per user
- [ ] Ban/Unban button per user
- [ ] Changes reflect immediately (query invalidation)

**Dependencies:** Task 1.3 (admin handlers)

---

### Task 3.3: System Health + Audit Log Pages (4h)

**Objective:** Create system health dashboard and audit log viewer.

**Files to Create:**
```
apps/frontend/src/routes/admin/health.tsx
apps/frontend/src/routes/admin/audit-log.tsx
```

**Implementation:**
```tsx
// routes/admin/health.tsx
import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { api } from '~/lib/api'
import { DashboardLayout } from '~/components/layout/DashboardLayout'
import { Card, CardContent, CardHeader, CardTitle } from '~/components/ui/card'

export const Route = createFileRoute('/admin/health')({
  component: SystemHealthPage,
})

function SystemHealthPage() {
  const { data: health, isLoading } = useQuery({
    queryKey: ['admin-health'],
    queryFn: () => api.get('/api/admin/health'),
    refetchInterval: 30000, // Refresh every 30s
  })
  
  if (isLoading) return <div>Loading...</div>
  
  return (
    <DashboardLayout>
      <div className="container mx-auto p-6">
        <h1 className="text-3xl font-bold mb-6">System Health</h1>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <Card>
            <CardHeader>
              <CardTitle>Cluster CPU</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{health.cluster_cpu}%</div>
              <Progress value={health.cluster_cpu} className="mt-2" />
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Cluster Memory</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{health.cluster_memory}%</div>
              <Progress value={health.cluster_memory} className="mt-2" />
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Cluster Storage</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{health.cluster_storage}%</div>
              <Progress value={health.cluster_storage} className="mt-2" />
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Total Users</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{health.total_users}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Total VMs</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{health.total_vms}</div>
            </CardContent>
          </Card>
          
          <Card>
            <CardHeader>
              <CardTitle>Active VMs</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-4xl font-bold">{health.active_vms}</div>
            </CardContent>
          </Card>
        </div>
      </div>
    </DashboardLayout>
  )
}
```

**Acceptance Criteria:**
- [ ] System health shows CPU, memory, storage gauges
- [ ] Auto-refresh every 30 seconds
- [ ] Audit log shows recent admin actions
- [ ] Audit log includes timestamp, user, action, IP

**Dependencies:** Task 1.3 (admin handlers)

---

### Task 3.4: Pin VM UI Integration (2h)

**Objective:** Add pin/unpin button to VM detail page.

**Files to Modify:**
```
apps/frontend/src/routes/dashboard/-vms/$id.tsx
```

**Implementation:**
```tsx
// routes/dashboard/-vms/$id.tsx
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '~/lib/api'

function VMDetailPage() {
  const { vm } = useVM()
  const queryClient = useQueryClient()
  
  const pinMutation = useMutation({
    mutationFn: () => api.post(`/api/vms/${vm.id}/pin`),
    onSuccess: () => {
      queryClient.invalidateQueries(['vm', vm.id])
    },
  })
  
  const unpinMutation = useMutation({
    mutationFn: () => api.del(`/api/vms/${vm.id}/pin`),
    onSuccess: () => {
      queryClient.invalidateQueries(['vm', vm.id])
    },
  })
  
  return (
    <DashboardLayout>
      <VMHeader vm={vm} />
      
      <div className="flex items-center gap-2 mt-4">
        {vm.is_pinned ? (
          <Button
            variant="outline"
            onClick={() => unpinMutation.mutate()}
          >
            <PinIcon className="w-4 h-4 mr-2" />
            Unpin VM
          </Button>
        ) : (
          <Button
            variant="outline"
            onClick={() => pinMutation.mutate()}
          >
            <PinIcon className="w-4 h-4 mr-2" />
            Pin VM (Prevent Auto-Delete)
          </Button>
        )}
        
        {vm.is_pinned && (
          <Badge variant="secondary">Pinned</Badge>
        )}
      </div>
      
      {/* Rest of VM detail */}
    </DashboardLayout>
  )
}
```

**Acceptance Criteria:**
- [ ] Pin button visible on VM detail page
- [ ] Button toggles between Pin/Unpin
- [ ] Pinned status shown with badge
- [ ] Pin limit enforced (error message if exceeded)

**Dependencies:** Task 2.4 (pin API endpoints)

---

## Week 4: Load Testing + Backup

### Task 4.1: k6 Load Testing Scripts (6h)

**Objective:** Create load testing scripts for critical user paths.

**Files to Create:**
```
tests/load/critical-paths.js
tests/load/README.md
```

**Implementation:**
```javascript
// tests/load/critical-paths.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metric for error rate
const errorRate = new Rate('errors');

export const options = {
  vus: 100,           // 100 concurrent users
  duration: '5m',     // 5 minutes
  thresholds: {
    http_req_duration: ['p(95)<500'], // p95 < 500ms
    http_req_failed: ['rate==0'],     // 0% errors
    errors: ['rate<0.01'],            // <1% custom errors
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  // 1. Auth flow (login)
  const loginRes = http.post(`${BASE_URL}/api/auth/login`);
  
  const loginSuccess = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login has token': (r) => r.json('access_token') !== '',
  });
  
  errorRate.add(!loginSuccess);
  
  if (!loginSuccess) {
    console.error('Login failed');
    return;
  }
  
  const token = loginRes.json('access_token');
  const headers = { 'Authorization': `Bearer ${token}` };
  
  sleep(1);
  
  // 2. VM create
  const createRes = http.post(
    `${BASE_URL}/api/vms`,
    JSON.stringify({ 
      name: `load-test-vm-${__VU}`, 
      os: 'ubuntu-2204', 
      tier: 'micro' 
    }),
    { headers }
  );
  
  check(createRes, {
    'create VM status is 201': (r) => r.status === 201,
    'create VM has id': (r) => r.json('id') !== '',
  });
  
  const vmID = createRes.json('id');
  sleep(1);
  
  // 3. VM start
  const startRes = http.post(
    `${BASE_URL}/api/vms/${vmID}/start`,
    null,
    { headers }
  );
  
  check(startRes, {
    'start VM status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // 4. Metrics fetch
  const metricsRes = http.get(
    `${BASE_URL}/api/vms/${vmID}/metrics?range=24h`,
    { headers }
  );
  
  check(metricsRes, {
    'metrics status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // 5. VM stop
  const stopRes = http.post(
    `${BASE_URL}/api/vms/${vmID}/stop`,
    null,
    { headers }
  );
  
  check(stopRes, {
    'stop VM status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
  
  // 6. VM delete
  const deleteRes = http.del(
    `${BASE_URL}/api/vms/${vmID}`,
    null,
    { headers }
  );
  
  check(deleteRes, {
    'delete VM status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
}
```

**Acceptance Criteria:**
- [ ] Script runs with k6
- [ ] 100 concurrent users for 5 minutes
- [ ] p95 response time < 500ms
- [ ] 0% error rate
- [ ] All critical paths tested (auth + VM lifecycle)

**Dependencies:** None (independent testing)

---

### Task 4.2: Load Test Execution (4h)

**Objective:** Run load tests and document results.

**Files to Create:**
```
.planning/phases/05-phase5/LOAD_TEST_RESULTS.md
```

**Execution Steps:**
```bash
# 1. Install k6
# macOS
brew install k6

# Windows (with Scoop)
scoop install k6

# 2. Run load test
k6 run tests/load/critical-paths.js

# 3. Run with custom URL
BASE_URL=http://podland-app.com k6 run tests/load/critical-paths.js

# 4. Export results
k6 run --out json=results.json tests/load/critical-paths.js
```

**Success Criteria:**
- [ ] p95 response time < 500ms (all endpoints)
- [ ] 0% error rate (no 5xx errors)
- [ ] 100 concurrent users sustained for 5 minutes
- [ ] All resources healthy (CPU < 80%, RAM < 80%)

**Acceptance Criteria:**
- [ ] Load test executed successfully
- [ ] Results documented in LOAD_TEST_RESULTS.md
- [ ] Any failures analyzed and fixed
- [ ] Retest after fixes

**Dependencies:** Task 4.1

---

### Task 4.3: Backup Automation (4h)

**Objective:** Create automated daily PostgreSQL backups.

**Files to Create:**
```
scripts/backup-db.sh
infra/k3s/backups/backup-cronjob.yaml
```

**Implementation:**
```bash
#!/bin/bash
# scripts/backup-db.sh

set -e

# Configuration
DB_NAME="podland"
DB_USER="podland"
DB_HOST="${DB_HOST:-localhost}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
S3_BUCKET="${S3_BUCKET:-}"
DATE=$(date +%Y%m%d-%H%M%S)

echo "Starting backup at $(date)"

# Create backup directory if not exists
mkdir -p "${BACKUP_DIR}"

# Create backup
echo "Creating backup..."
pg_dump "postgresql://${DB_USER}@${DB_HOST}/${DB_NAME}" | gzip > "${BACKUP_DIR}/podland-${DATE}.sql.gz"

BACKUP_SIZE=$(du -h "${BACKUP_DIR}/podland-${DATE}.sql.gz" | cut -f1)
echo "Backup created: podland-${DATE}.sql.gz (${BACKUP_SIZE})"

# Upload to S3 if configured
if [ -n "${S3_BUCKET}" ]; then
    echo "Uploading to S3..."
    aws s3 cp "${BACKUP_DIR}/podland-${DATE}.sql.gz" "${S3_BUCKET}/"
    echo "Upload complete"
fi

# Clean up old backups (keep 7 days)
echo "Cleaning up old backups..."
find "${BACKUP_DIR}" -name "podland-*.sql.gz" -mtime +7 -delete

echo "Backup completed successfully at $(date)"
```

**Kubernetes CronJob:**
```yaml
# infra/k3s/backups/backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: pg-backup
  namespace: podland
spec:
  schedule: "0 3 * * *"  # Daily at 3 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15
            command:
            - /scripts/backup-db.sh
            env:
            - name: DB_HOST
              value: "postgres"
            - name: DB_USER
              value: "podland"
            - name: DB_NAME
              value: "podland"
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: podland-secrets
                  key: database-password
            - name: S3_BUCKET
              value: "s3://podland-backups"
            volumeMounts:
            - name: backup-script
              mountPath: /scripts
            - name: backup-storage
              mountPath: /backups
          volumes:
          - name: backup-script
            configMap:
              name: backup-script
          - name: backup-storage
            persistentVolumeClaim:
              claimName: backup-pvc
          restartPolicy: OnFailure
```

**Acceptance Criteria:**
- [ ] Backup script runs successfully
- [ ] Backups created daily at 3 AM
- [ ] Backups uploaded to S3 (if configured)
- [ ] Old backups deleted after 7 days
- [ ] Backup restore tested (dry run)

**Dependencies:** None (independent infrastructure)

---

### Task 4.4: Documentation + Launch Prep (2h)

**Objective:** Create launch checklist and final documentation.

**Files to Create:**
```
.planning/phases/05-phase5/LAUNCH_CHECKLIST.md
```

**Implementation:**
```markdown
# Launch Checklist: Podland v1.0

## Pre-Launch (T-7 days)

### Infrastructure
- [ ] k3s cluster healthy
- [ ] PostgreSQL backup running (verify latest backup)
- [ ] Monitoring stack healthy (Prometheus, Loki, Grafana)
- [ ] Cloudflare Tunnel active

### Backend
- [ ] All migrations applied
- [ ] Environment variables set
- [ ] SendGrid API key configured
- [ ] Admin superadmin user created

### Frontend
- [ ] Build successful
- [ ] All routes accessible
- [ ] Mobile responsive tested

### Security
- [ ] Security audit passed (no critical/high vulnerabilities)
- [ ] Dependencies updated (npm audit, go mod verify)
- [ ] CSRF protection enabled
- [ ] Rate limiting configured

## Launch Day (T-0)

### Final Checks
- [ ] Load test passed (100 concurrent users)
- [ ] Database backup taken (manual pre-launch backup)
- [ ] Monitoring alerts configured
- [ ] On-call rotation established

### Deployment
- [ ] Deploy backend
- [ ] Deploy frontend
- [ ] Verify health endpoints
- [ ] Test critical user flows

### Post-Deployment
- [ ] Login flow works
- [ ] VM creation works
- [ ] Metrics visible
- [ ] Admin panel accessible
- [ ] Email notifications sent

## Post-Launch (T+7 days)

### Monitoring
- [ ] Review error logs daily
- [ ] Monitor resource usage
- [ ] Check idle detection working
- [ ] Verify backup restoration (dry run)

### User Feedback
- [ ] Collect user feedback
- [ ] Document issues encountered
- [ ] Plan Phase 5.1 improvements
```

**Acceptance Criteria:**
- [ ] Launch checklist complete
- [ ] All pre-launch items checked
- [ ] Deployment runbook documented
- [ ] Post-launch monitoring plan in place

**Dependencies:** All previous tasks

---

## Success Criteria Verification

| Requirement | Success Criteria | Verification Method |
|-------------|------------------|---------------------|
| ADMIN-01 | Superadmin can view list of all users | Admin panel shows user list with filters |
| ADMIN-02 | Superadmin can change user role | Role change persists, user sees new permissions |
| ADMIN-03 | Superadmin can ban/unban users | Banned user cannot sign in |
| ADMIN-04 | Superadmin can view system health dashboard | Dashboard shows cluster CPU/RAM/storage |
| ADMIN-05 | System logs all admin actions to audit log | Audit log shows who did what, when |
| IDLE-01 | System detects idle VMs (48h no activity) | Idle VMs identified hourly |
| IDLE-02 | System sends warning 24h before delete | User receives email + in-app notification |
| IDLE-03 | System auto-deletes idle VM after grace period | VM deleted if still idle after 24h |
| IDLE-04 | User can pin VM to prevent auto-delete | Pinned VMs excluded from idle detection |
| VM-08 | Quota enforcement verified at scale | Load test: 100 concurrent VMs, quotas enforced |

---

## Testing Checklist

### Backend Tests
```bash
# Run all tests
cd apps/backend
go test ./...

# Test admin handlers
go test ./internal/handler -run TestAdminHandler

# Test idle detector
go test ./internal/idle -run TestDetector
```

### Frontend Tests
```bash
# Run all tests
cd apps/frontend
npm test

# Run e2e tests
npm run test:e2e
```

### Load Testing
```bash
# Run load test
k6 run tests/load/critical-paths.js

# Run with JSON output
k6 run --out json=results.json tests/load/critical-paths.js
```

### Manual Testing
- [ ] Login as superadmin
- [ ] Access admin panel (/admin)
- [ ] View user list
- [ ] Change user role
- [ ] Ban/unban user
- [ ] View system health
- [ ] View audit log
- [ ] Pin VM
- [ ] Unpin VM
- [ ] Verify idle detection (create idle VM, wait for warning)

---

## Deployment Checklist

### Prerequisites
- [ ] k3s cluster running
- [ ] PostgreSQL database running
- [ ] SendGrid account created
- [ ] S3 bucket for backups created

### Backend Deployment
```bash
# Set environment variables
export SENDGRID_API_KEY=SG.xxxxx
export SENDGRID_FROM_EMAIL=noreply@podland.app
export PROMETHEUS_URL=http://prometheus.monitoring.svc:9090

# Run migrations
cd apps/backend
go run migrations/005_phase5_admin.sql

# Deploy
kubectl apply -f ../../infra/k3s/backend.yaml
```

### Frontend Deployment
```bash
# Build
cd apps/frontend
npm run build

# Deploy
kubectl apply -f ../../infra/k3s/frontend.yaml
```

### Backup Deployment
```bash
# Deploy backup CronJob
kubectl apply -f infra/k3s/backups/backup-cronjob.yaml
kubectl apply -f infra/k3s/backups/backup-pvc.yaml
```

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| False positive idle detection | Low | High | Combined criteria (HTTP + CPU) |
| Email delivery failures | Medium | Medium | Retry with backoff, in-app fallback |
| Load test failures | Medium | High | Strict criteria, optimize before launch |
| Backup corruption | Low | Critical | Test restore procedure |
| Admin panel XSS | Low | High | Input sanitization, CSP headers |

---

*Plan created: 2026-03-29*
*Ready for execution — follow week-by-week*
