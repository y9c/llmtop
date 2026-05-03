# llmtop

Real-time terminal dashboard for LLM inference servers.

Monitor vLLM (and other backends) GPU utilization, throughput, speculative decoding, prefix cache hit rate, and time-series charts — all in your terminal.

```text
llmtop ┃ vLLM ┃ NVIDIA L40S ┃ q
────────────────────────────────────────────────────────────────────────────────────────
┌─ GPU ────────────────────────────────────────────────────────────────────────────────┐
│ Memory  42.6GB / 43.9GB  ██████████████░  96.9%                                      │
│ Util                     ███████████████  97.0%                                      │
│ KV                         ░░░░░░░░░░░░░   9.3%                                      │
└──────────────────────────────────────────────────────────────────────────────────────┘
┌─ Throughput ─────────────────────┬─ Speculative ─────────────────────────────────────┐
│ run   1  wait   0                │ accept 48.6%  t/d 2.27                            │
│ dec 48.0  pre  0.0               │ draft 10.9K  rej 16.9K                           │
│ gen 35.5K  prm 153.5K            │ P0:8.3K P1:6.0K P2:4.4K P3:3.4K P4:2.6K          │
│                                  │ hit 60.0%  q 500.0K                              │
│                                  │ cache 358.8K  cmp 154.7K                         │
└──────────────────────────────────┴───────────────────────────────────────────────────┘
┌─ Timeline ───────────────────────────────────────────────────────────────────────────┐
│  48.00 ┤                              ╭──╮                                           │
│  36.00 ┤                    ╭────╮  ╭─╯  ╰─╮                                        │
│  24.00 ┤              ╭────╯    ╰──╯      ╰──╮                                      │
│  12.00 ┤         ╭────╯                      ╰────╮                                 │
│   0.00 ┼─────────╯                                ╰────                              │
│  97.21 ┼─────────────────────────────────────────────────                             │
└──────────────────────────────────────────────────────────────────────────────────────┘
 14:30:00 ┃ up 1h 23m ┃ gen 35.5K ┃ q
```

## Features

- **GPU monitoring** — VRAM usage, utilization %, KV cache usage with color-coded progress bars
- **Throughput** — running/waiting requests, decode/prefill tok/s, cumulative tokens
- **Speculative decoding** — accept rate, tokens/draft, per-position acceptance (P0–P4), draft/reject counts
- **Prefix cache** — hit rate %, queries, cached/computed tokens
- **Time-series charts** — asciigraph line charts for decode speed and memory usage
- **Multi-backend** — auto-detects vLLM, SGLang, llama.cpp from /metrics endpoint
- **Color-coded** — red ≥90%, yellow ≥70%, green <70% thresholds
- **Graceful shutdown** — Ctrl+C or `q` to quit, terminal restored

## Installation

### Prerequisites

- Go 1.24+
- `nvidia-smi` (for GPU monitoring)
- LLM inference server with Prometheus `/metrics` endpoint (vLLM, SGLang, llama.cpp)

### From source

```bash
git clone https://github.com/changye/llmtop.git
cd llmtop
make build
```

Or install directly:

```bash
go install github.com/changye/llmtop/cmd/llm-top@latest
```

### Pre-built binary

Download from [GitHub Releases](https://github.com/changye/llmtop/releases).

## Usage

```bash
# Connect to vLLM on localhost:8000 (default)
llmtop

# Specify port
llmtop --port 8080

# Specify host and port
llmtop --host 192.168.1.100 --port 8000

# Force a specific backend
llmtop --backend vllm

# Monitor a specific GPU
llmtop --gpu 0

# Faster update rate
llmtop --rate 500ms
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `localhost` | Metrics host |
| `--port` | `8000` | Metrics port |
| `--backend` | `auto` | Backend (`auto`, `vllm`, `sglang`, `ollama`, `llmcpp`) |
| `--rate` | `1s` | Update interval |
| `--gpu` | `-1` (all) | GPU ID (0-based) |
| `--help` | | Show help |
| `--version` | | Show version |

### Environment variables

All flags can also be set via environment variables: `LLM_TOP_HOST`, `LLM_TOP_PORT`, `LLM_TOP_BACKEND`, `LLM_TOP_RATE`, `LLM_TOP_GPU`.

### Controls

| Key | Action |
|-----|--------|
| `q` | Quit |
| `Ctrl+C` | Quit |

## Backends

| Backend | Detection | Status |
|---------|-----------|--------|
| **vLLM** | `vllm:` prefix in metrics | ✅ Full |
| **SGLang** | `sgl:` prefix | 🟡 Minimal |
| **llama.cpp** | `llm_prompt_tokens` or `slots_` prefix | 🟡 Minimal |

## Architecture

```text
llmtop/
├── cmd/llm-top/main.go       # Entry point (35 lines)
├── internal/
│   ├── app/app.go            # Orchestrator: ticker loop, fetch, delta compute
│   ├── backend/              # Backend interface + parsers (vLLM, SGLang, llama.cpp)
│   ├── collector/            # GPU collector (nvidia-smi)
│   ├── config/config.go      # CLI flags, env vars
│   ├── fetcher/fetcher.go    # HTTP client with retry + backoff
│   ├── metrics/              # Snapshot, Deltas, ring buffer, sparkline
│   └── ui/                   # bubbletea Model, View, styles, box-drawing helpers
├── Makefile
└── .github/workflows/release.yml
```

Built with [bubbletea](https://github.com/charmbracelet/bubbletea), [lipgloss](https://github.com/charmbracelet/lipgloss), [asciigraph](https://github.com/guptarohit/asciigraph).

## Development

```bash
make build     # Build binary
make run       # Build + run
make test      # Run all tests
make lint      # go vet
make clean     # Clean build artifacts
```

## License

MIT
