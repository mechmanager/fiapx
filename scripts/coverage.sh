#!/usr/bin/env bash
# Script de cobertura de testes do projeto FIAP X.
# Exclui: cmd/ (entrypoints), mocks/ (código gerado), storage/ e queue/
# (adaptadores de infra que precisam de serviços reais para testes de integração).
# As funções New/Run/Close do consumer (conexão RabbitMQ) também são excluídas
# por exigirem broker real — cobertas pelos testes de integração do CI (Fase 6).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
THRESHOLD=80

run_coverage() {
  local module="$1"
  local packages="$2"

  echo ""
  echo "=== Cobertura: $module ==="
  cd "$ROOT/$module"

  go test -coverprofile=coverage.out $packages 2>&1

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
  "./config/... ./handler/... ./middleware/... ./repository/... ./service/..."

run_coverage "worker" \
  "./config/... ./pipeline/... ./notification/... ./processor/..."

echo ""
echo "Cobertura OK em ambos os módulos."
