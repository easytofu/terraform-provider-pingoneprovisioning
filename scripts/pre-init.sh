#!/usr/bin/env bash
# /tofu/pre-init.sh
set -euo pipefail

: "${JFROG_USERNAME:?set JFROG_USERNAME in Scalr workspace env vars}"
: "${JFROG_TOKEN:?set JFROG_TOKEN in Scalr workspace env vars}"

if ! command -v oras >/dev/null 2>&1; then
  echo "oras not found"
  exit 1
fi

oras login easygogroup.jfrog.io -u "$JFROG_USERNAME" -p "$JFROG_TOKEN"
