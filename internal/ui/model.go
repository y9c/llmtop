package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/y9c/llmtop/internal/metrics"
)

type TickMsg struct {
	Backend  string
	GPUName  string
	Snap     metrics.Snapshot
	Delta    metrics.Deltas
	Uptime   time.Duration
	DecHist  []float64
	MemHist  []float64
	UtilHist []float64
	KVHist   []float64
}

type Model struct {
	Backend  string
	GPUName  string
	Snap     metrics.Snapshot
	Delta    metrics.Deltas
	Uptime   time.Duration
	DecHist  []float64
	MemHist  []float64
	UtilHist []float64
	KVHist   []float64
	Width    int
}

func (m Model) Init() tea.Cmd { return nil }

var _ tea.Model = (*Model)(nil)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
	case TickMsg:
		m.Backend = msg.Backend
		m.GPUName = msg.GPUName
		m.Snap = msg.Snap
		m.Delta = msg.Delta
		m.Uptime = msg.Uptime
		m.DecHist = msg.DecHist
		m.MemHist = msg.MemHist
		m.UtilHist = msg.UtilHist
		m.KVHist = msg.KVHist
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
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
