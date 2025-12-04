#!/usr/bin/env bash
set -euo pipefail

: "${MODEL_NAME:=gemma:2b}"
: "${OLLAMA_MODELS:=/models}"
: "${PORT:=8080}"

echo "[entrypoint] starting ollama serve"
ollama serve &

echo "[entrypoint] waiting for ollama..."
i=0; until curl -fsS http://127.0.0.1:11434/api/tags >/dev/null 2>&1; do
  i=$((i+1)); [ $i -gt 60 ] && { echo "ollama not ready"; exit 1; }
  sleep 1
done

echo "[entrypoint] pulling model: ${MODEL_NAME}"
ollama pull "${MODEL_NAME}" || true

sleep 2

echo "[entrypoint] starting flask on :${PORT}"
exec python3 /app/app.py
