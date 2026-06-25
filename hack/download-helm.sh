#!/bin/bash
set -euo pipefail

HELM_VERSION="${1:-v3.19.5}"
DEST_DIR="${2:-$(pwd)/bin/helm/${HELM_VERSION}}"

mkdir -p "${DEST_DIR}"

for arch in amd64 arm64; do
  target="${DEST_DIR}/helm-linux-${arch}"
  if [[ ! -f "${target}" ]]; then
    echo "Downloading helm ${HELM_VERSION} linux/${arch} ..."
    curl -fsSL "https://github.com/fluid-cloudnative/helm/releases/download/${HELM_VERSION}/helm-${HELM_VERSION}-linux-${arch}.tar.gz" \
      | tar -xz --strip-components=1 -C "${DEST_DIR}" "linux-${arch}/helm"
    mv "${DEST_DIR}/helm" "${target}"
    chmod +x "${target}"
  else
    echo "helm ${HELM_VERSION} linux/${arch} already exists, skipping."
  fi
done
