#!/usr/bin/env bash
set -euo pipefail

PKGS=(
  "./pkg/config"
  "./pkg/license"
  "./pkg/storage"
  "./pkg/llm"
  "./pkg/form"
  "./pkg/output"
)

echo "=== Running all package tests ==="
for pkg in "${PKGS[@]}"; do
  echo
  echo ">> Testing ${pkg}"
  go test -timeout 30s "${pkg}"
done

echo
echo "✅ All tests passed!"
