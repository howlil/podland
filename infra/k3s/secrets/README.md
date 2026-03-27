# Sealed Secrets Management

This directory contains SealedSecret manifests for managing secrets securely in k3s.

## What are Sealed Secrets?

Sealed Secrets is a Kubernetes controller and tool for one-way encryption of Secrets. Once sealed, the Secret can only be decrypted by the controller running in your cluster. This allows you to safely commit encrypted secrets to Git.

## Prerequisites

1. **Install kubeseal CLI:**
   ```bash
   # macOS
   brew install kubeseal
   
   # Windows (with scoop)
   scoop install kubeseal
   
   # Or download from: https://github.com/bitnami-labs/sealed-secrets/releases
   ```

2. **Install Sealed Secrets Controller in k3s:**
   ```bash
   kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/latest/download/controller.yaml
   ```

## Creating/Updating Secrets

### Step 1: Create a plain Secret YAML

**postgres-secret.yaml** (DO NOT COMMIT):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: podland
type: Opaque
stringData:
  password: "your-secure-password-here"
```

**podland-backend-secret.yaml** (DO NOT COMMIT):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: podland-backend-secret
  namespace: podland
type: Opaque
stringData:
  jwt-secret: "your-jwt-secret-min-32-chars"
  refresh-token-secret: "your-refresh-token-secret-min-32-chars"
  github-client-id: "your-github-client-id"
  github-client-secret: "your-github-client-secret"
```

### Step 2: Seal the Secret

```bash
# Seal individual secret
kubeseal --format yaml < postgres-secret.yaml > postgres-sealedsecret.yaml

# Or use raw mode for individual keys
echo -n "my-secure-password" | kubeseal --raw --from-file=/dev/stdin \
  --namespace podland --name postgres-secret --key password
```

### Step 3: Apply to Cluster

```bash
kubectl apply -f postgres-sealedsecret.yaml
kubectl apply -f backend-sealedsecret.yaml
```

### Step 4: Verify

```bash
kubectl get sealedsecret -n podland
kubectl get secret -n podland
```

## Cleanup

After sealing, delete the plain secret files:
```bash
rm postgres-secret.yaml podland-backend-secret.yaml
```

## Generating Secure Passwords

```bash
# Generate secure random password (32 chars)
openssl rand -base64 32

# Or using Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"
```

## Files

| File | Description |
|------|-------------|
| `postgres-sealedsecret.yaml` | Encrypted PostgreSQL credentials |
| `backend-sealedsecret.yaml` | Encrypted backend API credentials |
| `README.md` | This documentation |

## Troubleshooting

**Sealed secret not becoming ready:**
- Ensure Sealed Secrets controller is running: `kubectl get pods -n kube-system -l name=sealed-secrets-controller`
- Check controller logs: `kubectl logs -n kube-system -l name=sealed-secrets-controller`

**Namespace mismatch:**
- Sealed secrets are namespace-scoped by default
- Ensure `--namespace` matches your target namespace
