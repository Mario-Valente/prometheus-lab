# Prometheus Lab

Lab prático de Prometheus com Kind, kube-prometheus-stack e aplicação Go exemplo.

## Pré-requisitos

- Docker
- kubectl
- Helm 3
- K6 (opcional, para load test)

## Quick Start

1. Clone e entre no diretório:
   git clone <repo>
   cd prometheus-lab

2. Inicie o ambiente:
   ./scripts/start.sh

3. Valide a instalação:
   ./scripts/validate.sh

4. Acesse as interfaces (abra 3 terminais):
   kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 &
   kubectl port-forward svc/prometheus-grafana 3000:80 &
   kubectl port-forward svc/prometheus-app 8080:8080 &

5. Acesse:
   Prometheus: http://localhost:9090
   Grafana: http://localhost:3000 (admin/admin)
   App Go: http://localhost:8080/metrics

## Estrutura do Repositório

kind-config.yaml         - Configuração do Kind cluster
prometheus-values.yaml   - Valores Helm do kube-prometheus-stack
app/                     - Aplicação Go que expõe métricas
scripts/
  start.sh              - Cria cluster e instala stack
  validate.sh           - Valida todo o setup
  cleanup.sh            - Remove tudo
  load-test.js          - Script K6 para gerar carga

## Componentes Instalados

Prometheus Server       - Scraper e TSDB de séries temporais
Grafana               - Visualização de métricas
Alertmanager          - Gerenciamento de alertas
Kube-state-metrics    - Métricas do Kubernetes
Node-exporter         - Métricas de hardware
Aplicação Go          - App exemplo que expõe métricas

## Uso da Aplicação Go

Endpoints:
GET /metrics           - Prometheus scrape endpoint
GET /health            - Health check
POST /request/:id      - Simula uma requisição
GET /connections       - Status de conexões

Métricas expostas:
http_requests_total           - Contador de requisições
active_connections            - Número de conexões ativas
request_duration_seconds      - Histograma de latência

## Load Test

Para gerar carga e ver métricas em ação:

  k6 run scripts/load-test.js

Ou execute via Prometheus:
  kubectl port-forward svc/prometheus-app 8080:8080
  k6 run --vus 50 --duration 2m scripts/load-test.js

## Troubleshooting

Se algum Pod não subir:
  kubectl describe pod <pod-name>
  kubectl logs <pod-name>

Para ver targets do Prometheus:
  kubectl logs deployment/prometheus | grep target

Para limpar tudo:
  ./scripts/cleanup.sh

## Referências

- Prometheus: https://prometheus.io/docs
- Kube-prometheus-stack: https://github.com/prometheus-community/helm-charts
- PromQL: https://prometheus.io/docs/prometheus/latest/querying/basics

## Nota de Segurança

IMPORTANTE: Este ambiente de laboratório utiliza credenciais padrão e não deve ser exposto publicamente.

- Credenciais Grafana padrão: admin/admin
- O Grafana solicitará a alteração de senha no primeiro login
- Para ambientes de produção, sempre configure credenciais seguras e autenticação adequada
- Não exponha os serviços port-forwarded para redes públicas
