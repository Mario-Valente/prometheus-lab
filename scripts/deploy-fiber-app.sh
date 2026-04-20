#!/bin/bash

# Script para build e deploy do Fiber Prometheus App

set -e

echo "🚀 Building and deploying Fiber Prometheus App..."

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para log
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1"
}

# Verificar se estamos no diretório correto
if [[ ! -f "fiber-app/go.mod" ]]; then
    error "Arquivo fiber-app/go.mod não encontrado. Execute este script do diretório raiz do projeto."
    exit 1
fi

# Nome do cluster Kind (padrão: prometheus-lab)
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-prometheus-lab}"

# Build da aplicação
log "Building Go application..."
cd fiber-app
go mod download
go mod tidy

# Build da imagem Docker
log "Building Docker image..."
docker build -t fiber-prometheus-app:latest .
success "Docker image built successfully"

# Voltar para o diretório raiz
cd ..

# Verificar se o Kind cluster está rodando
log "Checking Kind cluster..."
if ! kind get clusters | grep -q "^${KIND_CLUSTER_NAME}$"; then
    error "Kind cluster '${KIND_CLUSTER_NAME}' não encontrado. Execute ./scripts/start.sh primeiro."
    exit 1
fi
success "Kind cluster '${KIND_CLUSTER_NAME}' is running"

# Carregar imagem no Kind
log "Loading image into Kind cluster..."
kind load docker-image --name "${KIND_CLUSTER_NAME}" fiber-prometheus-app:latest
success "Image loaded into Kind cluster '${KIND_CLUSTER_NAME}'"

# Deploy do ServiceMonitor primeiro (se existir o CRD)
log "Deploying ServiceMonitor and PrometheusRule..."
if kubectl get crd servicemonitors.monitoring.coreos.com &>/dev/null; then
    kubectl apply -f k8s/fiber-app-servicemonitor.yaml
    success "ServiceMonitor and PrometheusRule deployed"
else
    warning "ServiceMonitor CRD not found. Make sure Prometheus Operator is installed."
fi

# Deploy da aplicação
log "Deploying Fiber Prometheus App..."
kubectl apply -f k8s/fiber-app-deployment.yaml
success "Fiber app deployment applied"

# Aguardar deployment estar pronto
log "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/fiber-prometheus-app
success "Deployment is ready"

# Mostrar informações do deployment
log "Deployment information:"
kubectl get pods -l app=fiber-prometheus-app
kubectl get service fiber-prometheus-app

# Port forward para acesso local (opcional)
echo ""
log "To access the application locally, run:"
echo "  kubectl port-forward service/fiber-prometheus-app 8081:8080"
echo ""
log "Then access:"
echo "  Health: http://localhost:8081/health"
echo "  Metrics: http://localhost:8081/metrics"
echo "  API Users: http://localhost:8081/api/v1/users"

success "Fiber Prometheus App deployed successfully! 🎉"