# Stack Research: PaaS Platform 2025

## Recommended Stack for Podland

### Backend

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Language** | Go 1.21+ | High performance, excellent for concurrent operations, strong Kubernetes ecosystem support, minimal resource footprint |
| **Web Framework** | Gin or Echo | Lightweight, high-performance, excellent middleware support |
| **Authentication** | OAuth2 (GitHub) + JWT | GitHub OAuth required, JWT for session management |
| **Database** | PostgreSQL 16 | ACID compliance, JSONB for flexible schemas, excellent Go drivers |
| **Cache** | Redis 7 | Session storage, rate limiting, job queues |
| **API Style** | REST + gRPC | REST for external API, gRPC for internal service communication |

### Frontend

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Framework** | TanStack Start 1.0+ | Full-stack React framework, file-based routing, server-side rendering, type-safe |
| **UI Library** | Tailwind CSS v4 | Utility-first, excellent for dashboards, v4 has improved performance |
| **State Management** | Zustand | Lightweight, simple API, perfect for dashboard state |
| **Data Fetching** | Axios + TanStack Query | Caching, background updates, optimistic updates |
| **Monorepo** | Turborepo + pnpm | Fast builds, efficient caching, workspace management |

### Infrastructure

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Orchestration** | k3s 1.29+ | Lightweight Kubernetes, production-ready, 500MB RAM footprint, perfect for single-server |
| **Package Manager** | Helm 3.14+ | Chart-based deployments, version control, rollback support |
| **Container Runtime** | containerd | Default in k3s, lighter than Docker daemon |
| **Ingress** | Traefik 2.x | Default in k3s, automatic Let's Encrypt, simple configuration |
| **Service Mesh** | Linkerd (optional) | Lightweight mTLS, observability, can add later |

### Observability

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **Metrics** | Prometheus + Grafana | Industry standard, excellent k3s integration, rich ecosystem |
| **Logging** | Loki + Promtail | Lightweight, integrates with Grafana, cost-effective |
| **Tracing** | Tempo (optional) | Distributed tracing, can add in v2 |
| **Alerting** | Alertmanager | Native Prometheus integration, routing rules |

### Storage

| Type | Technology | Use Case |
|------|------------|----------|
| **local** | Host filesystem | OS, binaries, temporary data, container images |
| **local-lvm** | LVM-backed PV | Persistent volumes for user data, databases |
| **NFS (optional)** | Longhorn | Distributed block storage, replication |

### DevOps

| Component | Technology | Rationale |
|-----------|------------|-----------|
| **IaC** | Terraform + Kubernetes provider | VM provisioning, Cloudflare DNS/Tunnel automation |
| **CI/CD** | GitHub Actions | Native GitHub integration, free for open source |
| **Registry** | Harbor or Docker Hub | Container image storage, vulnerability scanning |

## What NOT to Use

| Technology | Why Avoid |
|------------|-----------|
| **Full Kubernetes** | Overkill for single-server, k3s is CNCF-certified and lighter |
| **Docker Swarm** | Limited ecosystem, declining adoption, poor observability integration |
| **Nomad** | Different use case, less cloud-native tooling |
| **Istio** | Too heavy for single-server, Linkerd is lighter if needed |
| **Elasticsearch** | Resource hog, Loki sufficient for logging at this scale |
| **MongoDB** | Schema-less adds complexity, PostgreSQL JSONB sufficient |
| **RabbitMQ** | Redis streams sufficient for job queues at this scale |

## Version Notes (Verified 2025)

- Go: 1.21+ (stable), 1.22 (beta)
- k3s: 1.29.x (current stable)
- TanStack Start: 1.0+ (released late 2024)
- Tailwind CSS: 4.0 (released 2025)
- Turborepo: 2.0+ (Vercel maintained)
- Helm: 3.14+
- PostgreSQL: 16.x (17 in beta)

## Installation Strategy

```bash
# k3s single-node installation
curl -sfL https://get.k3s.io | sh -

# Verify cluster
kubectl get nodes
kubectl get pods --all-namespaces

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Add Helm repos
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```
