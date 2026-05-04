#!/bin/bash
# Start Ollama server
set -euo pipefail

PORT="${1:-11434}"

echo "=== Starting Ollama server ==="
echo "   Port: $PORT"
echo ""

# Ollama binary is expected in PATH or at ~/.local/bin/ollama
OLLAMA_BIN="${OLLAMA_BIN:-ollama}"
if ! command -v "$OLLAMA_BIN" &>/dev/null; then
    echo "Ollama not found in PATH. Install: curl -fsSL https://ollama.com/install.sh | sh"
    exit 1
fi

export OLLAMA_HOST="0.0.0.0:$PORT"
export OLLAMA_KEEP_ALIVE="5m"

exec "$OLLAMA_BIN" serve
