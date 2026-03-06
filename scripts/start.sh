#!/bin/bash
set -e

echo "===== PROMETHEUS LAB - STARTUP ====="

# Load environment
if [ -f .env ]; then
  export $(cat .env | xargs)
else
  echo ".env not found, using defaults"
fi

KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-prometheus-lab}
KIND_IMAGE=${KIND_IMAGE:-kindest/node:v1.29.0}
PROMETHEUS_RETENTION=${PROMETHEUS_RETENTION:-30d}
GRAFANA_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}

# Step 1: Create Kind cluster
echo ""
echo "[1/7] Creating Kind cluster..."
if kind get clusters | grep -q "^${KIND_CLUSTER_NAME}$"; then
  echo "Kind cluster '${KIND_CLUSTER_NAME}' already exists, skipping creation"
else
  kind create cluster --name "${KIND_CLUSTER_NAME}" --config k8s/kind-config.yaml
fi

# Set kubectl context
kubectl cluster-info --context kind-${KIND_CLUSTER_NAME}

# Step 2: Add Helm repositories
echo ""
echo "[2/7] Adding Helm repositories..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Step 3: Install kube-prometheus-stack
echo ""
echo "[3/7] Installing kube-prometheus-stack via Helm..."
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
  --values k8s/prometheus-values.yaml \
  --wait \
  --timeout 5m

# Step 4: Build Go app image
echo ""
echo "[4/7] Building Go application Docker image..."
docker build -t prometheus-app:latest app/

# Step 5: Load image into Kind cluster
echo ""
echo "[5/7] Loading image into Kind cluster..."
kind load docker-image prometheus-app:latest --name "${KIND_CLUSTER_NAME}"

# Step 6: Deploy application
echo ""
echo "[6/7] Deploying application..."
kubectl apply -f k8s/app-deployment.yaml

# Step 7: Wait for rollout
echo ""
echo "[7/7] Waiting for Deployments to be Ready..."
kubectl rollout status deployment/prometheus-app --timeout=2m
kubectl rollout status deployment/prometheus-kube-prometheus-prometheus --timeout=2m
kubectl rollout status deployment/prometheus-grafana --timeout=2m

echo ""
echo "===== SETUP COMPLETE ====="
echo "Run: ./scripts/validate.sh"
