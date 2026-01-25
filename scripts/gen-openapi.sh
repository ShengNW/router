#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

SWAG_BIN="$(command -v swag || true)"
if [ -z "$SWAG_BIN" ]; then
  SWAG_BIN="$(go env GOPATH)/bin/swag"
fi

"$SWAG_BIN" init --generalInfo docs/swagger.go --output docs --parseDependency --parseInternal

if [ -f docs/swagger.json ]; then
  cp docs/swagger.json docs/openapi.json
fi
