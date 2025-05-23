#!/bin/bash
set -e

KUSTOMIZATION_FILE="deploy/kustomization.yaml"
NEXT_VERSION="$1"
if [ -z "$NEXT_VERSION" ]; then
  echo "Usage: $0 <next-version>"
  exit 1
fi

sed -i "s/ref=[^ ]*/ref=${NEXT_VERSION}/" "$KUSTOMIZATION_FILE"
sed -i "s/newTag: [^ ]*/newTag: ${NEXT_VERSION}/" "$KUSTOMIZATION_FILE"