#!/bin/bash

echo "===== PROMETHEUS LAB - CLEANUP ====="
echo "This will remove the Kind cluster and all associated resources"
echo ""

read -p "Are you sure? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Cleanup cancelled"
  exit 1
fi

KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-prometheus-lab}

echo "Removing Kind cluster: $KIND_CLUSTER_NAME"
kind delete cluster --name "${KIND_CLUSTER_NAME}"

echo "Removing local Docker image..."
docker rmi prometheus-app:latest 2>/dev/null || true

echo ""
echo "===== CLEANUP COMPLETE ====="
