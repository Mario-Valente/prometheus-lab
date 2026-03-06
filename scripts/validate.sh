#!/bin/bash
set -e

echo "===== PROMETHEUS LAB - VALIDATION ====="
echo ""

FAILED=0
PASSED=0

check() {
  local name=$1
  local command=$2
  local expected=$3

  echo -n "Checking: $name ... "

  if output=$(eval "$command" 2>&1); then
    if [[ "$output" == *"$expected"* ]]; then
      echo "✅ PASS"
      ((PASSED++))
      return 0
    fi
  fi

  echo "❌ FAIL"
  ((FAILED++))
  return 1
}

# Checks
echo "[INFRASTRUCTURE]"
check "Kind cluster exists" "kind get clusters | grep -c prometheus-lab" "1"
check "kubectl context" "kubectl config current-context" "kind-prometheus-lab"

echo ""
echo "[PROMETHEUS STACK]"
check "Prometheus pod running" "kubectl get pods -l app.kubernetes.io/name=prometheus -o jsonpath='{.items[0].status.phase}'" "Running"
check "Grafana pod running" "kubectl get pods -l app.kubernetes.io/name=grafana -o jsonpath='{.items[0].status.phase}'" "Running"
check "Alertmanager pod running" "kubectl get pods -l app.kubernetes.io/name=alertmanager -o jsonpath='{.items[0].status.phase}'" "Running"
check "Kube-state-metrics running" "kubectl get pods -l app.kubernetes.io/name=kube-state-metrics -o jsonpath='{.items[0].status.phase}'" "Running"
check "Node-exporter running" "kubectl get pods -l app.kubernetes.io/name=node-exporter -o jsonpath='{.items[0].status.phase}'" "Running"

echo ""
echo "[APPLICATION]"
check "App pod running" "kubectl get pods -l app=prometheus-app -o jsonpath='{.items[0].status.phase}'" "Running"
check "App service exists" "kubectl get svc prometheus-app -o jsonpath='{.metadata.name}'" "prometheus-app"

echo ""
echo "[METRICS COLLECTION]"
# Temporarily port-forward to test metrics endpoint
kubectl port-forward svc/prometheus-app 8080:8080 >/dev/null 2>&1 &
PF_PID=$!
sleep 1

check "App metrics endpoint" "curl -s http://localhost:8080/metrics | grep -c http_requests_total" "1"
check "Counter metric exists" "curl -s http://localhost:8080/metrics | grep -c active_connections" "1"
check "Histogram metric exists" "curl -s http://localhost:8080/metrics | grep -c request_duration_seconds" "1"

kill $PF_PID 2>/dev/null || true
wait $PF_PID 2>/dev/null || true

echo ""
echo "[INTEGRATION]"
check "Prometheus targets up" "kubectl logs -l app.kubernetes.io/name=prometheus -n default --tail=100 | grep -c 'scrape_configs'" "1"

echo ""
echo "===== VALIDATION SUMMARY ====="
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo ""

if [ $FAILED -eq 0 ]; then
  echo "✅ ALL CHECKS PASSED"
  echo ""
  echo "Next steps:"
  echo "1. Port-forward Prometheus:   kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 &"
  echo "2. Port-forward Grafana:      kubectl port-forward svc/prometheus-grafana 3000:80 &"
  echo "3. Port-forward App:          kubectl port-forward svc/prometheus-app 8080:8080 &"
  echo ""
  echo "Then access:"
  echo "- Prometheus: http://localhost:9090"
  echo "- Grafana: http://localhost:3000 (admin/admin)"
  echo "- App Metrics: http://localhost:8080/metrics"
  exit 0
else
  echo "❌ SOME CHECKS FAILED"
  echo ""
  echo "Troubleshooting:"
  echo "kubectl describe pod <pod-name>"
  echo "kubectl logs <pod-name>"
  exit 1
fi
