#!/usr/bin/env bash
# Cobertura de testes de todos os módulos Go do projeto FIAP X.
# Exclui cmd/ (entrypoints), mocks/ e adaptadores de infra (precisam de serviços reais).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
THRESHOLD=80

run_coverage() {
  local module="$1"
  shift
  local packages=("$@")

  echo ""
  echo "=== Cobertura: $module ==="
  cd "$ROOT/$module"

  go test -coverprofile=coverage.out "${packages[@]}" 2>&1

  local total
  total=$(go tool cover -func=coverage.out | awk '/^total:/ {gsub(/%/,""); print $3}')
  echo "Total ($module): ${total}%"

  if awk "BEGIN { exit !($total >= $THRESHOLD) }"; then
    echo "OK — acima de ${THRESHOLD}%"
  else
    echo "FALHA — abaixo de ${THRESHOLD}% (obteve ${total}%)"
    exit 1
  fi
}

run_coverage "api-gateway" \
  "./config/..." "./middleware/..." "./proxy/..."

run_coverage "auth-service" \
  "./config/..." "./domain/..." "./service/..." "./handler/..."

run_coverage "upload-service" \
  "./config/..." "./domain/..." "./service/..." "./handler/..."

run_coverage "status-service" \
  "./config/..." "./domain/..." "./service/..." "./handler/..."

run_coverage "notification-service" \
  "./config/..." "./domain/..." "./mailer/..."

run_coverage "worker" \
  "./config/..." "./pipeline/..." "./notification/..." "./processor/..."

echo ""
echo "Cobertura OK em todos os módulos."
