#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"

export HF_TOKEN="$(cat "$ROOT_DIR/.hf_token" 2>/dev/null)"
export PYTORCH_CUDA_ALLOC_CONF="expandable_segments:True"
export VLLM_USE_V1=0
export VLLM_ALLOW_LONG_MAX_MODEL_LEN=1
export VLLM_ENGINE_READY_TIMEOUT_S=1800

source "$ROOT_DIR/.venv/bin/activate"

echo "Starting vLLM..."
vllm serve Qwen/Qwen3.6-27B-FP8 \
  --trust-remote-code --port 8000 --host 0.0.0.0 \
  --max-model-len 262144 --gpu-memory-utilization 0.933 \
  --kv-cache-dtype fp8 --served-model-name qwen3.6 \
  --language-model-only --max-num-seqs 8 --max-num-batched-tokens 8192 \
  --enable-chunked-prefill --enable-prefix-caching \
  --api-key "sk-test" \
  --speculative-config '{"method":"mtp","num_speculative_tokens":2}' \
  --reasoning-parser qwen3 --tool-call-parser qwen3_coder
