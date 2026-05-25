#!/usr/bin/env bash
# Build and load all AWX execution environments.
# Run on the host where AWX is deployed.
# Requires: ansible-builder (pip install ansible-builder)
# Images are pushed to localhost:5000 (or REGISTRY if set).

set -euo pipefail

REGISTRY="${REGISTRY:-localhost/home-platform}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EE_DIR="$SCRIPT_DIR/execution-environments"

build_ee() {
  local name="$1"
  local dir="$EE_DIR/$name"
  local tag="$REGISTRY/$name:latest"

  echo "==> Building $name → $tag"
  ansible-builder build \
    --file "$dir/execution-environment.yml" \
    --tag "$tag" \
    --container-runtime docker \
    --verbosity 2

  echo "==> Loading $tag into Docker"
  docker tag "$tag" "$tag"
  echo "Done: $name"
}

build_ee base-ee
build_ee docker-ee
build_ee k8s-ee

echo ""
echo "All EEs built. Register them in AWX via:"
echo "  ansible-playbook awx_config/configure_awx.yml --tags ee"
