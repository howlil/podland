#!/bin/bash
# Script untuk generate Sealed Secrets
# Usage: ./generate-sealed-secrets.sh

set -e

NAMESPACE="podland"
SECRETS_DIR="secrets"

echo "🔐 Sealed Secrets Generator"
echo "==========================="
echo ""

# Check kubectl connection
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Cannot connect to Kubernetes cluster!"
    echo "   Make sure k3s/k3d is running and kubectl is configured"
    exit 1
fi

echo "✅ Connected to cluster"

# Check if Sealed Secrets controller is running
if ! kubectl get pods -n kube-system -l name=sealed-secrets-controller &> /dev/null; then
    echo "⚠️  Sealed Secrets controller not found!"
    echo "   Install with: kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/latest/download/controller.yaml"
    exit 1
fi

echo "✅ Sealed Secrets controller is running"
echo ""

# Create secrets directory
mkdir -p "$SECRETS_DIR"

# Generate PostgreSQL password
read -p "Enter PostgreSQL password (or press Enter to auto-generate): " POSTGRES_PASSWORD
if [ -z "$POSTGRES_PASSWORD" ]; then
    POSTGRES_PASSWORD=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)
    echo "📝 Generated PostgreSQL password: $POSTGRES_PASSWORD"
fi

# Generate JWT secret
read -p "Enter JWT secret (min 32 chars, or press Enter to auto-generate): " JWT_SECRET
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)
    echo "📝 Generated JWT secret: $JWT_SECRET"
fi

# Generate Refresh Token secret
read -p "Enter Refresh Token secret (min 32 chars, or press Enter to auto-generate): " REFRESH_SECRET
if [ -z "$REFRESH_SECRET" ]; then
    REFRESH_SECRET=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)
    echo "📝 Generated Refresh Token secret: $REFRESH_SECRET"
fi

# GitHub credentials
read -p "Enter GitHub Client ID: " GITHUB_CLIENT_ID
read -s -p "Enter GitHub Client Secret: " GITHUB_CLIENT_SECRET
echo ""

echo ""
echo "🔒 Creating plain secrets (temporary)..."

# Create PostgreSQL secret
cat > /tmp/postgres-secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secret
  namespace: $NAMESPACE
type: Opaque
stringData:
  password: "$POSTGRES_PASSWORD"
EOF

# Create Backend secret
cat > /tmp/backend-secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: podland-backend-secret
  namespace: $NAMESPACE
type: Opaque
stringData:
  jwt-secret: "$JWT_SECRET"
  refresh-token-secret: "$REFRESH_SECRET"
  github-client-id: "$GITHUB_CLIENT_ID"
  github-client-secret: "$GITHUB_CLIENT_SECRET"
EOF

echo "✅ Plain secrets created"
echo ""
echo "🔐 Sealing secrets..."

# Seal PostgreSQL secret
kubeseal --format yaml < /tmp/postgres-secret.yaml > "$SECRETS_DIR/postgres-sealedsecret.yaml"
echo "✅ Sealed PostgreSQL secret"

# Seal Backend secret
kubeseal --format yaml < /tmp/backend-secret.yaml > "$SECRETS_DIR/backend-sealedsecret.yaml"
echo "✅ Sealed Backend secret"

# Cleanup plain secrets
rm -f /tmp/postgres-secret.yaml /tmp/backend-secret.yaml
echo "🗑️  Cleaned up plain secrets"

echo ""
echo "✅ Sealed secrets generated successfully!"
echo ""
echo "📁 Files created:"
echo "   - $SECRETS_DIR/postgres-sealedsecret.yaml"
echo "   - $SECRETS_DIR/backend-sealedsecret.yaml"
echo ""
echo "📝 Next steps:"
echo "   1. Review the sealed secret files"
echo "   2. Apply to cluster: kubectl apply -f $SECRETS_DIR/"
echo "   3. Commit sealed secrets to git (safe to commit!)"
echo ""
echo "⚠️  IMPORTANT: Save these values somewhere safe!"
echo "   PostgreSQL Password: $POSTGRES_PASSWORD"
echo "   JWT Secret: $JWT_SECRET"
echo "   Refresh Token Secret: $REFRESH_SECRET"
