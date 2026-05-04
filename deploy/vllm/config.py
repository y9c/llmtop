"""vLLM deployment configuration for Qwen3.6-27B-FP8 on L40S."""

from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent.parent
HF_TOKEN = (ROOT / ".hf_token").read_text().strip()

class VLLMConfig:
    # Model
    model: str = "Qwen/Qwen3.6-27B-FP8"
    served_name: str = "qwen3.6"

    # Server
    host: str = "0.0.0.0"
    port: int = 8000
    api_key: str = "sk-8f861cd49f01a5a9a575fefa90ef678d425652ad3474c14d"

    # GPU / Memory
    gpu_memory_util: float = 0.933
    max_model_len: int = 262144
    kv_cache_dtype: str = "fp8"
    tensor_parallel: int = 1

    # Scheduling
    max_num_seqs: int = 16
    max_num_batched_tokens: int = 12288
    enable_chunked_prefill: bool = True
    enable_prefix_caching: bool = True

    # Speculative Decoding (MTP)
    speculative_config: str = '{"method":"mtp","num_speculative_tokens":2}'

    # Features
    language_model_only: bool = True
    reasoning_parser: str = "qwen3"
    tool_call_parser: str = "qwen3_coder"

    # Logging
    disable_log_requests: bool = True

    def cli_args(self) -> list[str]:
        flags = [
            self.model,
            "--trust-remote-code",
            "--tensor-parallel-size", str(self.tensor_parallel),
            "--port", str(self.port),
            "--host", self.host,
            "--max-model-len", str(self.max_model_len),
            "--gpu-memory-utilization", str(self.gpu_memory_util),
            "--kv-cache-dtype", self.kv_cache_dtype,
            "--served-model-name", self.served_name,
            "--speculative-config", self.speculative_config,
            "--max-num-seqs", str(self.max_num_seqs),
            "--max-num-batched-tokens", str(self.max_num_batched_tokens),
            "--reasoning-parser", self.reasoning_parser,
            "--tool-call-parser", self.tool_call_parser,
            "--api-key", self.api_key,
        ]
        if self.language_model_only:
            flags.append("--language-model-only")
        if self.enable_chunked_prefill:
            flags.append("--enable-chunked-prefill")
        if self.enable_prefix_caching:
            flags.append("--enable-prefix-caching")
        if self.disable_log_requests:
            flags.append("--disable-log-requests")
        return flags
