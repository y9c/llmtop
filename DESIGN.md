# llmtop — Real-time LLM Inference Monitor (TUI)

Terminal dashboard for LLM inference servers. Built with [bubbletea](https://github.com/charmbracelet/bubbletea), [lipgloss](https://github.com/charmbracelet/lipgloss), [asciigraph](https://github.com/guptarohit/asciigraph).

## Architecture

```
llmtop/
├── cmd/llm-top/main.go       # Entry point (35 lines)
├── internal/
│   ├── app/app.go            # Orchestrator: ticker loop, fetch, delta compute
│   ├── backend/              # Backend interface + parsers (vLLM, SGLang, llama.cpp)
│   ├── collector/            # GPU collector (nvidia-smi)
│   ├── config/config.go      # CLI flags, env vars
│   ├── fetcher/fetcher.go    # HTTP client with retry + backoff
│   ├── metrics/              # Snapshot, Deltas, ring buffer
│   └── ui/                   # bubbletea Model, View, styles, box drawing
├── Makefile
├── DESIGN.md
└── .github/workflows/release.yml
```

## Data Flow

```
[ticker (1s)] ─→ [nvidia-smi] + [http.Get /metrics]
                    ↓
          [backend.Parse()] → Snapshot
                    ↓
        [program.Send(TickMsg)]
            ├── compute delta vs prev snapshot
            ├── push to ring buffers (60s history)
            └── buildView() renders full UI
```

## Key Types

```go
type Snapshot struct {
    Timestamp                               time.Time
    GPUName                                 string
    GPUUsedMB, GPUMemTotalMB, GPUUtilPct    float64
    KVCacheUsagePct                         float64
    RunningReqs, WaitingReqs                float64
    GenTokensTotal, PromptTokensTotal        float64
    PromptCachedTotal, PromptLocalTotal      float64
    SpecDraftsTotal, SpecDraftToksTotal      float64
    SpecAcceptedTotal                        float64
    SpecAcceptedPos                         []float64
    PrefixCacheHits, PrefixCacheQueries      float64
    IterTimeTotalS, IterTimeCount            float64
    TPOTTotalS, TPOTCount                   float64
    TTFTTotalS, TTFTCount                   float64
}

type Deltas struct {
    DecodeTokS, PrefillTokS, AcceptRate     float64
}
```

## UI Layout

Box-drawing borders with sections for GPU, Throughput, Speculative Decoding, and Timeline.

```
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
```

## Regex Patterns (vLLM)

| Metric | Pattern |
|--------|---------|
| `RunningReqs` | `num_requests_running\{[^}]*\}\s+([\d.eE+-]+)` |
| `WaitingReqs` | `num_requests_waiting_by_reason\{[^}]*reason="capacity"[^}]*\}\s+([\d.eE+-]+)` |
| `KVCacheUsagePct` | `kv_cache_usage_perc\{[^}]*\}\s+([\d.eE+-]+)` |
| `GenTokensTotal` | `generation_tokens_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `PromptTokensTotal` | `prompt_tokens_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `PromptCachedTotal` | `prompt_tokens_cached_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `PromptLocalTotal` | `prompt_tokens_by_source_total\{[^}]*source="local_compute"[^}]*\}\s+([\d.eE+-]+)` |
| `SpecDraftsTotal` | `spec_decode_num_drafts_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `SpecDraftToksTotal` | `spec_decode_num_draft_tokens_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `SpecAcceptedTotal` | `spec_decode_num_accepted_tokens_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `SpecAcceptedPos[i]` | `spec_decode_num_accepted_tokens_per_pos_total\{[^}]*position="{i}"[^}]*\}\s+([\d.eE+-]+)` |
| `PrefixCacheHits` | `prefix_cache_hits_total\{[^}]*\}\s+([\d.eE+-]+)` |
| `PrefixCacheQueries` | `prefix_cache_queries_total\{[^}]*\}\s+([\d.eE+-]+)` |

## History & Sparklines

- **Ring buffer**: 60 data points (1 per second, 60s window)
- **Sparklines**: 8-level Unicode `▁▂▃▄▅▆▇█`

## Backend Interface

```go
type Backend interface {
    Name() string
    Detect(body string) bool
    Parse(body string) (metrics.Snapshot, error)
}
```

## Dependencies

- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — Styling
- `github.com/guptarohit/asciigraph` — Time-series charts
- `nvidia-smi` (external) — GPU data
