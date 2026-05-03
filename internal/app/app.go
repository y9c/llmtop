package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"log"

	"github.com/changye/llmtop/internal/backend"
	"github.com/changye/llmtop/internal/collector"
	"github.com/changye/llmtop/internal/config"
	"github.com/changye/llmtop/internal/fetcher"
	"github.com/changye/llmtop/internal/metrics"
	"github.com/changye/llmtop/internal/ui"
)

type App struct {
	cfg     *config.Config
	fetcher *fetcher.Fetcher
	gpu     collector.GPUCollector
	program *tea.Program
	model   *ui.Model

	memHist *metrics.History
	kvHist  *metrics.History
	utilHist *metrics.History
	decHist *metrics.History

	// Reusable buffers to avoid heap alloc per tick
	decBuf  []float64
	memBuf  []float64
	utilBuf []float64
	kvBuf   []float64

	prevSnap metrics.Snapshot
	prevSet  bool
	backend  backend.Backend
	startAt  time.Time
	gpuName  string
}

func New(cfg *config.Config, f *fetcher.Fetcher, gpu collector.GPUCollector, m *ui.Model) *App {
	return &App{
		cfg:     cfg,
		fetcher: f,
		gpu:     gpu,
		model:    m,
		memHist:  metrics.NewHistory(),
		kvHist:   metrics.NewHistory(),
		utilHist: metrics.NewHistory(),
		decHist:  metrics.NewHistory(),
		startAt:  time.Now(),
	}
}

func (a *App) Run(ctx context.Context) error {
	a.program = tea.NewProgram(a.model, tea.WithAltScreen())

	go func() {
		<-ctx.Done()
		a.program.Quit()
	}()

	go a.tick(ctx)

	_, err := a.program.Run()
	return err
}



func (a *App) tick(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.Rate)
	defer ticker.Stop()

	// Fire first tick immediately
	a.doFetch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.doFetch(ctx)
		}
	}
}

func (a *App) doFetch(ctx context.Context) {
	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	gpuInfo, gpuErr := a.gpu.Fetch(fetchCtx)
	if gpuErr == nil && a.gpuName == "" {
		a.gpuName = gpuInfo.Name
	}

	body, httpErr := a.fetcher.Fetch(fetchCtx, a.cfg.MetricsURL())

	if httpErr == nil && a.backend == nil {
		a.backend = backend.Detect(body)
	}
	if a.backend == nil {
		a.backend = &backend.VLLM{}
	}

	var snap metrics.Snapshot
	if httpErr == nil {
		var err error
		snap, err = a.backend.Parse(body)
		if err != nil {
			log.Printf("parse metrics: %v", err)
		}
	}
	snap.Timestamp = time.Now()

	if gpuErr == nil {
		snap.GPUUsedMB = gpuInfo.UsedMB
		snap.GPUMemTotalMB = gpuInfo.TotalMB
		snap.GPUUtilPct = gpuInfo.UtilPct
		snap.GPUName = gpuInfo.Name
	}

	var delta metrics.Deltas
	if a.prevSet {
		delta = metrics.ComputeDelta(a.prevSnap, snap, a.cfg.Rate.Seconds())
	}
	a.prevSnap = snap
	a.prevSet = true

	if gpuErr == nil && snap.GPUMemTotalMB > 0 {
		a.memHist.Push(snap.GPUUsedMB / snap.GPUMemTotalMB * 100)
	}
	a.kvHist.Push(snap.KVCacheUsagePct * 100)
	a.utilHist.Push(snap.GPUUtilPct)
	a.decHist.Push(delta.DecodeTokS)

	a.program.Send(ui.TickMsg{
		Backend: a.backend.Name(),
		GPUName: a.gpuName,
		Snap:    snap,
		Delta:   delta,
		Uptime:  time.Since(a.startAt),
		DecHist: a.decHist.ValuesInto(a.decBuf),
		MemHist: a.memHist.ValuesInto(a.memBuf),
		UtilHist: a.utilHist.ValuesInto(a.utilBuf),
		KVHist:   a.kvHist.ValuesInto(a.kvBuf),
	})
}
