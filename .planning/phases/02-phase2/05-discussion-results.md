# Phase 2 Discussion Results

**Session:** 2026-03-27
**Status:** In Progress (1/4 decisions made)

---

## VM Image Strategy

### Q1: Base Image ✅ DECIDED
**Decision:** Option B — Build sendiri dari official images

**Rationale:**
- Best UX for students (tools ready immediately)
- Consistent environment across all VMs
- Security: base from official Ubuntu/Debian
- Size: ~300MB (fast pull <30s)
- Matches VPS provider pattern (DigitalOcean, etc.)

**Implementation:**
```dockerfile
FROM ubuntu:22.04

# Add common tools students need
RUN apt-get update && apt-get install -y \
    git \
    curl \
    wget \
    vim \
    htop \
    net-tools \
    iputils-ping \
    ca-certificates \
    gnupg \
    && rm -rf /var/lib/apt/lists/*

# Non-root user (security)
RUN useradd -m -u 1000 -s /bin/bash vmuser
USER vmuser
WORKDIR /home/vmuser

CMD ["/bin/bash"]
```

---

## Pending Decisions

### Q2: Preinstall Packages ✅ DECIDED
**Decision:** Option A — Essentials only

**Rationale:**
- Best UX balance: common tools ready, user chooses runtime versions
- Fast image pull (~300MB, <30s)
- Security: minimal attack surface
- Flexible: users install their own Node.js/Python/Go versions via nvm/pyenv/gvm

**Packages included:**
```bash
git, curl, wget, vim, htop, net-tools, iputils-ping, ca-certificates, gnupg
```

**User workflow:**
```bash
# Need Node.js?
nvm install --lts    # 30 seconds

# Need Python?
pyenv install 3.11   # 2 minutes

# Need Go?
gvm install go1.21   # 1 minute
```

### Q3: Build & Hosting ✅ DECIDED
**Decision:** Option A — GitHub Actions + Docker Hub (Official OS images only)

**Rationale:**
- VPS provider pattern: start with clean OS images only
- Simple: we maintain OS base, users install anything they want
- Low maintenance: no app images, no marketplace complexity
- Flexible: users can install Docker, WordPress, Node.js, etc. themselves

**Phase 2 OS Catalog:**
```
Operating Systems:
├── Ubuntu 22.04 LTS (recommended)
└── Debian 12
```

**Build pipeline:**
```yaml
# GitHub Actions monthly builds
podland/ubuntu-2204:latest  → Docker Hub
podland/debian-12:latest    → Docker Hub
```

**User workflow:**
```
1. User selects OS (Ubuntu 22.04 or Debian 12)
2. User selects tier (nano → xlarge)
3. VM created (pulls from Docker Hub)
4. User SSH in → install anything they want
```

**Future (Phase 5):**
- App images (WordPress, Docker, LAMP)
- Marketplace (user-submitted images)
- More OS options (Rocky Linux, Alpine, Windows)

### Q4: Versioning & Updates ✅ DECIDED
**Decision:** Option A — Semantic versioning + monthly builds

**Rationale:**
- VPS provider standard (DigitalOcean, Vultr, Linode pattern)
- Clear version history
- Monthly security patches (automated via GitHub Actions)
- User can pin version for reproducibility
- Existing VMs don't auto-update (user chooses when to recreate)

**Image tags:**
```
podland/ubuntu-2204:latest      → Always latest
podland/ubuntu-2204:v1.0        → January 2026 build
podland/ubuntu-2204:v1.1        → February 2026 build
podland/ubuntu-2204:2026-03     → March 2026 build
```

**User workflow:**
```bash
# Create VM with latest
POST /api/vms
{ "os": "ubuntu-2204", "version": "latest" }

# Create VM with pinned version (reproducible)
POST /api/vms
{ "os": "ubuntu-2204", "version": "v1.0" }

# Existing VMs don't auto-update
# User recreates VM if they want newer image
```

**GitHub Actions schedule:**
```yaml
on:
  schedule:
    - cron: '0 0 1 * *'  # 1st of every month
```

---

## Summary: VM Image Strategy (All Decided ✅)

| Decision | Choice |
|----------|--------|
| Base Image | Build sendiri dari official Ubuntu/Debian |
| Preinstall | Essentials only (git, curl, wget, vim, htop) |
| Build & Hosting | GitHub Actions + Docker Hub (OS only) |
| Versioning | Semantic + monthly (v1.0, v1.1, 2026-03) |

**Next:** Discuss VM Storage/Partitioning strategy...

---

## VM Storage Strategy

### Q5.1: Storage Type & Mounting ✅ DECIDED
**Decision:** Option A — Single root filesystem (`/` = 100% quota)

**Rationale:**
- Standard VPS pattern (DigitalOcean, Vultr, Linode all use single `/`)
- Simple: user familiar, no partition confusion
- Tier already limits: nano=5GB, micro=10GB, etc. includes OS + data
- k8s simplicity: single PVC per VM, no multi-mount complexity

**User experience:**
```bash
# User SSH in
ssh user@my-vm.podland.app

# Check disk usage (standard VPS experience)
df -h
# Filesystem      Size  Used Avail
# /dev/sda1       10G   2G   8G   ← nano tier = 10GB total

# User install stuff, no partition worries
apt install nginx
mkdir /var/www/myapp
```

**Implementation:**
```yaml
# Single PVC mounted at /
volumeMounts:
- name: vm-storage
  mountPath: /
  
# PVC size matches tier
spec:
  resources:
    requests:
      storage: 10Gi  # nano tier
```

### Pending: Q5.2: Storage Class & Persistence

### Q5.2: Storage Class & Persistence ✅ DECIDED
**Decision:** Option B — local-lvm storage class

**Rationale:**
- Matches Phase 1 decision (PROJECT.md: "local-lvm for persistent volumes")
- Better than plain local: LVM allows snapshots, better management
- VPS provider pattern: real VPS providers use LVM
- Perfect for single bare-metal server

**Storage classes:**
```yaml
# infra/k3s/storage-classes.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-lvm
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
```

**VM PVC uses local-lvm:**
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: vm-{vm-id}-pvc
  namespace: user-{user-id}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi  # matches tier
  storageClassName: local-lvm
```

**Persistence behavior:**
| Event | Storage |
|-------|---------|
| Container restart | ✅ Preserved |
| VM stop/start | ✅ Preserved |
| VM delete | ❌ Deleted (PVC deleted) |
| VM recreate | ❌ New PVC (user loses data) |

**Note:** User perlu backup sendiri (rsync, scp, dll) sebelum delete VM.

### Pending: Q5.3: Backup Strategy

### Q5.3: Backup Strategy ✅ DECIDED
**Decision:** Option B — Snapshot on VM delete (7-day retention)

**Rationale:**
- VPS provider balance (DigitalOcean/Vultr pattern)
- Safety net: user accidental delete, can recover within 7 days
- Manageable cost: snapshot only on delete, not periodic
- Auto-cleanup: snapshot auto-delete after 7 days

**Implementation:**
```yaml
# When user deletes VM:
# 1. Create snapshot of PVC
# 2. Keep for 7 days
# 3. Auto-delete after 7 days

# Snapshot naming
vm-{vm-id}-pvc-snapshot-{timestamp}

# User can restore within 7 days:
POST /api/vms/{id}/restore
```

**User workflow:**
```bash
# User delete VM (accidentally or on purpose)
DELETE /api/vms/vm-123

# System creates snapshot
# Notification: "VM deleted. Snapshot kept until 2026-04-03"

# User changes mind within 7 days:
POST /api/vms/vm-123/restore
# VM recreated with all data

# After 7 days:
# Snapshot auto-deleted, can't recover
```

**Storage overhead:**
- Snapshot only on delete (max 2 VMs per external user)
- 7 day retention = manageable storage
- Can add quota: "2 snapshots per user"

### Pending: Q5.4: OS Disk Size Calculation

### Q5.4: OS Disk Size Calculation ✅ DECIDED
**Decision:** Option A — Total includes OS (standard VPS model)

**Rationale:**
- Standard VPS model (DigitalOcean, Vultr, Linode all use this)
- Transparent: user knows what they get (5GB total)
- Predictable: no complex accounting
- Teaches resource management: user learns to manage disk like real VPS

**Tier breakdown (Total = OS + User):**

| Tier | Total | OS (Ubuntu) | Usable |
|------|-------|-------------|---------|
| nano | 5 GB | ~2.5 GB | ~2.5 GB |
| micro | 10 GB | ~2.5 GB | ~7.5 GB |
| small | 20 GB | ~2.5 GB | ~17.5 GB |
| medium | 40 GB | ~2.5 GB | ~37.5 GB |
| large | 80 GB | ~2.5 GB | ~77.5 GB |
| xlarge | 100 GB | ~2.5 GB | ~97.5 GB |

**User experience:**
```bash
# User SSH in, check disk
df -h
# Filesystem      Size  Used Avail
# /dev/sda1       5.0G  2.5G  2.5G  ← nano tier

# User sees actual usable space
# Clear expectation: they have 5GB total
```

**Note:** OS size stable (~2.5GB) because no auto-update. User updates their own OS.

---

## Summary: VM Storage Strategy (All Decided ✅)

| Decision | Choice |
|----------|--------|
| Storage Type | Single root filesystem (`/` = 100% quota) |
| Storage Class | local-lvm (matches Phase 1 decision) |
| Backup Strategy | Snapshot on VM delete (7-day retention) |
| OS Disk Calculation | Total includes OS (standard VPS model) |

**Next:** Discuss VM Networking, Security, or other gray areas...

---

## VM Network Access Strategy

### Q6.1: SSH Access Method ✅ DECIDED
**Decision:** Option A — SSH via Ingress (`ssh user@vm-name.podland.app`)

**Rationale:**
- VPS provider standard (DigitalOcean, Vultr, Linode all use SSH)
- User expectation: students expect `ssh user@vm-name.podland.app`
- Tool compatibility: scp, rsync, sftp, VS Code Remote all need SSH
- Simpler implementation: Traefik Ingress + SSH Service per VM

**Implementation:**
```yaml
# VM Service exposes SSH port
apiVersion: v1
kind: Service
metadata:
  name: vm-{vm-id}-ssh
  namespace: user-{user-id}
spec:
  selector:
    app: vm-{vm-id}
  ports:
  - name: ssh
    port: 22
    targetPort: 22
  type: ClusterIP

# Traefik IngressRoute for SSH (TCP)
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: vm-{vm-id}-ssh
  namespace: user-{user-id}
spec:
  entryPoints:
    - ssh  # Port 22 on host
  routes:
    - match: HostSNI(`vm-name.podland.app`)
      services:
        - name: vm-{vm-id}-ssh
          port: 22
```

**User workflow:**
```bash
# User create VM
POST /api/vms
{ "name": "my-server", "tier": "micro" }

# VM ready, user gets SSH key
# Download private key or use web-generated key

# User SSH in
ssh -i ~/.ssh/podland my-server@my-server.podland.app

# User can also use scp, rsync, sftp
scp file.txt my-server@my-server.podland.app:/home/vmuser/
```

**SSH Key Strategy:**
- Option: User uploads public key during VM creation
- Option: System generates keypair, user downloads private key
- Store public key in VM `~/.ssh/authorized_keys`

### Pending: Q6.2: SSH Key Management

### Q6.2: SSH Key Management ✅ DECIDED
**Decision:** Option B — System generates keypair (Ed25519)

**Rationale:**
- Target user: students (many don't have SSH keys yet)
- Zero friction: create VM → SSH immediately
- VPS provider trend: DigitalOcean, Vultr all auto-generate
- Security manageable: private key shown once, backend doesn't store it

**Implementation:**
```go
// Backend generates Ed25519 keypair
privateKey, publicKey, err := ssh.GenerateKeyPair()

// Store public key in database
vm.SSHPublicKey = publicKey

// Inject into VM (cloud-init or startup script)
#cloud-config
users:
  - name: vmuser
    ssh_authorized_keys:
      - {{publicKey}}

// Show private key to user ONCE (in UI)
-----BEGIN OPENSSH PRIVATE KEY-----
...
-----END OPENSSH PRIVATE KEY-----

[Download] [Copy] ⚠️ Show only once!
```

**User workflow:**
```
1. User create VM
2. VM created (30s)
3. UI shows: "VM ready! Download your SSH key"
4. User downloads `podland-my-server.pem`
5. User sets permission: `chmod 600 podland-my-server.pem`
6. User SSH: `ssh -i podland-my-server.pem vmuser@my-server.podland.app`
```

**Security:**
- Private key shown only once (like AWS keys)
- Backend stores only public key
- User responsible for securing private key
- Option: User can regenerate key later (via UI)

### Pending: Q6.3: VM Port Exposure

### Q6.3: VM Port Exposure ✅ DECIDED
**Decision:** Option A — All ports exposed via Ingress (flexible VPS model)

**Rationale:**
- VPS flexibility: user can run anything (web server, game server, bot, database)
- Common use cases covered:
  - Web app → port 80/443
  - Development server → port 3000, 8080, etc.
  - Game server (Minecraft) → port 25565
  - Discord bot → no inbound needed
  - Database (for testing) → port 5432

**Implementation:**
```yaml
# User creates VM, wants to expose port 3000 (Node.js app)
# System creates IngressRouteTCP

apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: vm-{vm-id}-http
  namespace: user-{user-id}
spec:
  entryPoints:
    - web      # Port 80
    - websecure # Port 443
  routes:
    - match: HostSNI(`vm-name.podland.app`)
      services:
        - name: vm-{vm-id}-http
          port: 80

# User can also expose custom ports
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRouteTCP
metadata:
  name: vm-{vm-id}-custom
spec:
  entryPoints:
    - game     # Port 25565 (Minecraft)
  routes:
    - match: HostSNI(`vm-name.podland.app`)
      services:
        - name: vm-{vm-id}-game
          port: 25565
```

**Port mapping:**
```
External                    → VM Internal
vm-name.podland.app:80      → VM:80
vm-name.podland.app:443     → VM:443
vm-name.podland.app:3000    → VM:3000
vm-name.podland.app:8080    → VM:8080
vm-name.podland.app:25565   → VM:25565 (Minecraft)
```

**User workflow:**
```bash
# User SSH in
ssh vmuser@vm-name.podland.app

# User starts Node.js app on port 3000
node app.js --port 3000

# App accessible at:
# http://vm-name.podland.app:3000
```

**Security:**
- Firewall rules per VM (default deny, user opens ports)
- Rate limiting via Traefik
- DDoS protection via Cloudflare (Phase 3)

### Pending: Q6.4: Firewall Rules

### Q6.4: Firewall Rules ✅ DECIDED
**Decision:** Option B — Default allow common ports (22, 80, 443)

**Rationale:**
- Best UX for students: VM ready, SSH + web app works immediately
- VPS provider pattern: DigitalOcean, Vultr, Linode all open 22/80/443 by default
- Covers 90% use cases:
  - Port 22: SSH access (required)
  - Port 80: HTTP web app / redirect to HTTPS
  - Port 443: HTTPS web app
- User can add more ports: via UI or inside VM (ufw/iptables)

**Default firewall rules:**
```yaml
# Default inbound rules (applied to all VMs)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: vm-default-firewall
  namespace: user-{user-id}
spec:
  podSelector:
    matchLabels:
      app: vm
  policyTypes:
  - Ingress
  ingress:
  # Allow SSH (port 22)
  - ports:
    - protocol: TCP
      port: 22
  # Allow HTTP (port 80)
  - ports:
    - protocol: TCP
      port: 80
  # Allow HTTPS (port 443)
  - ports:
    - protocol: TCP
      port: 443
  # Egress: Allow all (for apt updates, etc.)
  egress:
  - {}
```

**User workflow:**
```
1. User create VM
2. VM ready
3. Default: SSH (22), HTTP (80), HTTPS (443) open
4. User SSH in: `ssh vmuser@vm-name.podland.app`
5. User deploys web app on port 3000
6. User wants to expose port 3000:
   - Option A: UI → "Add Port" → 3000
   - Option B: Inside VM, configure ufw: `ufw allow 3000`
```

**UI for port management (Phase 2 or 5):**
```
VM: my-server
─────────────────────────────────────
Open Ports:
  ✅ Port 22 (SSH)     [Edit] [Close]
  ✅ Port 80 (HTTP)    [Edit] [Close]
  ✅ Port 443 (HTTPS)  [Edit] [Close]
  ✅ Port 3000 (Custom) [Edit] [Close]

[+ Add Port]
```

**Security notes:**
- Default allow 22/80/443 (secure enough for most cases)
- User can add more ports via UI
- User can configure ufw/iptables inside VM for fine-grained control
- Rate limiting via Traefik (prevent DDoS)

---

## Summary: VM Network Access Strategy (All Decided ✅)

| Decision | Choice |
|----------|--------|
| SSH Access | SSH via Ingress (`ssh user@vm-name.podland.app`) |
| SSH Key Management | System generates keypair (Ed25519) |
| Port Exposure | All ports via Ingress |
| Firewall Rules | Default allow 22/80/443, user can add more |

**Next:** Discuss VM Security, Quota Enforcement, or API Design...

---

*Last updated: 2026-03-27 — VM Image, Storage & Network decisions complete*
