# llmtop

Real-time terminal dashboard for LLM inference servers. Monitor vLLM (and other backends) GPU utilization, throughput, speculative decoding, prefix cache, and time-series charts.

![screenshot](./docs/screenshot.png)

## Installation

```bash
curl -sfL https://raw.githubusercontent.com/changye/llmtop/master/install.sh | sh
```

Or build from source:

```bash
git clone https://github.com/changye/llmtop.git
cd llmtop && make build
```

## Usage

```bash
llmtop                          # Connect to localhost:8000
llmtop --port 8080              # Different port
llmtop --host 192.168.1.100     # Remote host
llmtop --rate 500ms             # Faster updates
```

| Flag | Default | Description |
|------|---------|-------------|
| `--host` | `localhost` | Metrics host |
| `--port` | `8000` | Metrics port |
| `--backend` | `auto` | Force backend (`vllm`, `sglang`, `ollama`) |
| `--rate` | `1s` | Update interval |
| `--gpu` | `-1` (all) | GPU ID |
| `q` / `Ctrl+C` | | Quit |

## Architecture

```
llmtop/
├── cmd/llm-top/main.go
├── internal/
│   ├── app/           # Ticker loop, fetch, delta compute
│   ├── backend/       # vLLM, SGLang, llama.cpp parsers
│   ├── collector/     # nvidia-smi GPU data
│   ├── config/        # CLI flags and env vars
│   ├── fetcher/       # HTTP client with retry
│   ├── metrics/       # Ring buffer + snapshot types
│   └── ui/            # bubbletea TUI renderer
├── Makefile
└── install.sh
```

Built with [bubbletea](https://github.com/charmbracelet/bubbletea), [lipgloss](https://github.com/charmbracelet/lipgloss), [asciigraph](https://github.com/guptarohit/asciigraph).

## License

MIT
