#!/usr/bin/env bash
# Install or verify ingress-nginx controller for k3d/k3s clusters.
# Idempotent: safe to run when ingress-nginx is already installed.
set -euo pipefail

INGRESS_NGINX_VERSION="v1.10.1"

# Apply the manifest — kubectl apply is a no-op when resources are unchanged.
echo "Applying ingress-nginx ${INGRESS_NGINX_VERSION} manifests..."
kubectl apply -f "https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-${INGRESS_NGINX_VERSION}/deploy/static/provider/cloud/deploy.yaml"

echo "Waiting for ingress-nginx controller to be ready (timeout 120s)..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=120s

echo "ingress-nginx is ready."
