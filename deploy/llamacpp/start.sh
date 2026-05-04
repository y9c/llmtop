#!/bin/bash
# Start llama.cpp server with uv-managed environment
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
UV="$HOME/.local/bin/uv"

if [ ! -x "$UV" ]; then
    echo "UV not found at $UV"
    exit 1
fi

MODEL="${1:-Qwen/Qwen3.6-27B-FP8}"
PORT=8001

echo "=== Starting llama.cpp server ==="
echo "   Model: $MODEL | Port: $PORT"
echo ""

cd "$ROOT_DIR"
exec "$UV" run --with llama-cpp-python --with "uvicorn[standard]" \
  python3 -m llama_cpp.server \
  --model "$MODEL" \
  --host 0.0.0.0 \
  --port "$PORT" \
  --n_gpu_layers -1 \
  --chat_format chatml
