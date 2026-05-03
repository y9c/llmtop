# llmtop

Real-time terminal dashboard for LLM inference servers. Monitor GPU utilization, throughput, speculative decoding, prefix cache, and timeline charts — all in your terminal.

![screenshot](./docs/llmtop_view.png)

## Installation

```bash
curl -sfL https://raw.githubusercontent.com/y9c/llmtop/main/install.sh | sh
```

Or build from source:

```bash
git clone https://github.com/y9c/llmtop.git
cd llmtop && make build
```

## Usage

```bash
llmtop                          # Connect to localhost:8000 (default)
llmtop --port 8080              # Different port
llmtop --host 192.168.1.100     # Remote host
llmtop --rate 500ms             # Faster updates
llmtop --gpu 0                  # Monitor specific GPU
llmtop --url http://ollama:11434/api/ps  # Custom metrics URL
```

| Flag | Default | Description |
|------|---------|-------------|
| `--url` | `""` | Full metrics URL (overrides host/port) |
| `--host` | `localhost` | Metrics host |
| `--port` | `8000` | Metrics port |
| `--backend` | `auto` | Force backend (`vllm`, `sglang`, `ollama`) |
| `--rate` | `1s` | Update interval |
| `--gpu` | `-1` (all) | GPU ID (0-based) |

`q` or `Ctrl+C` to quit.

## Backends

| Backend | Status |
|---------|--------|
| **vLLM** | Full metrics |
| **SGLang** | Basic |
| **llama.cpp** | Basic |
| **Ollama** | Basic |

## Features

- GPU temperature (°C) and power draw (W) in footer
- Session-wide min/avg/max TTFT and TPOT latency tracking
- Custom metrics URL via `--url` flag

## License

MIT
