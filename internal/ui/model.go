package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/y9c/llmtop/internal/metrics"
)

type LatencyStats struct {
	TTFTMinMs  float64
	TTFTAvgMs  float64
	TTFTMaxMs  float64
	TPOTMinMs  float64
	TPOTAvgMs  float64
	TPOTMaxMs  float64
}

type TickMsg struct {
	Backend  string
	GPUName  string
	Snap     metrics.Snapshot
	Delta    metrics.Deltas
	Uptime   time.Duration
	DecHist  []float64
	PreHist  []float64
	MemHist  []float64
	UtilHist []float64
	KVHist   []float64
	Latency  LatencyStats
}

type Model struct {
	Backend  string
	GPUName  string
	Snap     metrics.Snapshot
	Delta    metrics.Deltas
	Uptime   time.Duration
	DecHist  []float64
	PreHist  []float64
	MemHist  []float64
	UtilHist []float64
	KVHist   []float64
	Latency  LatencyStats
	Width    int
	Height   int
	Scroll   int // viewport scroll offset (lines from top)
}

func (m Model) chartHeight() int {
	// Overhead: title(1) + sep(1) + tp table(top1+rows4+bot1) + chart block(top1+gap1+bot1) ≈ 11
	// Remaining rows split into 2 chart blocks (Util+KV and Dec+Mem), each block gets half.
	avail := m.Height - 11
	if avail < 4 {
		return 0 // no room for charts
	}
	h := avail / 2
	if h > 6 {
		h = 6
	}
	return h
}

func (m Model) Init() tea.Cmd { return nil }

var _ tea.Model = (*Model)(nil)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case TickMsg:
		m.Backend = msg.Backend
		m.GPUName = msg.GPUName
		m.Snap = msg.Snap
		m.Delta = msg.Delta
		m.Uptime = msg.Uptime
		m.DecHist = msg.DecHist
		m.PreHist = msg.PreHist
		m.MemHist = msg.MemHist
		m.UtilHist = msg.UtilHist
		m.KVHist = msg.KVHist
		m.Latency = msg.Latency
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.Scroll > 0 {
				m.Scroll--
			}
		case "down", "j":
			m.Scroll++
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.Backend == "" || m.GPUName == "" {
		return "connecting..."
	}
	return m.buildView()
}
