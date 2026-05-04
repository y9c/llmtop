#!/bin/bash
# Start SGLang server with uv-managed environment
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
UV="$HOME/.local/bin/uv"

if [ ! -x "$UV" ]; then
    echo "UV not found at $UV"
    exit 1
fi

MODEL="${1:-Qwen/Qwen3.6-27B-FP8}"
PORT=8002

echo "=== Starting SGLang server ==="
echo "   Model: $MODEL | Port: $PORT"
echo ""

cd "$ROOT_DIR"
exec "$UV" run --with sglang --with "outlines" \
  python3 -m sglang.launch_server \
  --model-path "$MODEL" \
  --host 0.0.0.0 \
  --port "$PORT" \
  --trust-remote-code \
  --tp 1 \
  --context-length 32768 \
  --disable-log-requests
