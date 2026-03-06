#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: ./scripts/logs.sh <pod-pattern>"
  echo "Examples:"
  echo "  ./scripts/logs.sh prometheus"
  echo "  ./scripts/logs.sh grafana"
  echo "  ./scripts/logs.sh prometheus-app"
  exit 1
fi

kubectl logs -f -l $(kubectl get pods -o json | jq -r ".items[] | select(.metadata.name | contains(\"$1\")) | .metadata.labels | to_entries | map(\"\(.key)=\(.value)\") | .[0]" | head -1)
