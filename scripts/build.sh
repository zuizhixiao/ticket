#!/usr/bin/env bash
# 一键构建:UI → 单二进制(等价 scripts/build.ps1)
set -euo pipefail
cd "$(dirname "$0")/.."

if [[ "${SKIP_UI:-0}" != "1" ]]; then
  echo "[1/2] 构建前端..."
  (cd web && npm install --ignore-scripts && npm run build)
else
  echo "[1/2] 跳过前端构建(SKIP_UI=1)"
fi

echo "[2/2] 编译后端二进制 ticket..."
CGO_ENABLED=0 go build -o ticket ./cmd/server
echo "完成: $(pwd)/ticket"
