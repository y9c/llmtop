package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"log"

	"github.com/y9c/llmtop/internal/backend"
	"github.com/y9c/llmtop/internal/collector"
	"github.com/y9c/llmtop/internal/config"
	"github.com/y9c/llmtop/internal/fetcher"
	"github.com/y9c/llmtop/internal/metrics"
	"github.com/y9c/llmtop/internal/ui"
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
	preHist *metrics.History

	// Reusable buffers to avoid heap alloc per tick
	decBuf  []float64
	memBuf  []float64
	utilBuf []float64
	kvBuf   []float64
	preBuf  []float64

	prevSnap metrics.Snapshot
	prevSet  bool
	backend  backend.Backend
	startAt  time.Time
	gpuName  string

	// Session-wide latency tracking
	ttftMinS  float64
	ttftMaxS  float64
	ttftSumS  float64
	ttftN     float64
	tpotMinS  float64
	tpotMaxS  float64
	tpotSumS  float64
	tpotN     float64
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
		preHist:  metrics.NewHistory(),
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
		snap.GPUTempC = gpuInfo.TempC
		snap.GPUPowerW = gpuInfo.PowerW
		snap.GPUPowerMaxW = gpuInfo.PowerMaxW
		snap.GPUName = gpuInfo.Name
	}

	var delta metrics.Deltas
	if a.prevSet {
		delta = metrics.ComputeDelta(a.prevSnap, snap, a.cfg.Rate.Seconds())

		// Track per-sample TTFT/TPOT from histogram deltas
		if ttftN := snap.TTFTCount - a.prevSnap.TTFTCount; ttftN > 0 {
			ttftS := (snap.TTFTTotalS - a.prevSnap.TTFTTotalS) / ttftN
			if a.ttftN == 0 {
				a.ttftMinS = ttftS
				a.ttftMaxS = ttftS
			} else {
				if ttftS < a.ttftMinS { a.ttftMinS = ttftS }
				if ttftS > a.ttftMaxS { a.ttftMaxS = ttftS }
			}
			a.ttftSumS += ttftS * ttftN
			a.ttftN += ttftN
		}
		if tpotN := snap.TPOTCount - a.prevSnap.TPOTCount; tpotN > 0 {
			tpotS := (snap.TPOTTotalS - a.prevSnap.TPOTTotalS) / tpotN
			if a.tpotN == 0 {
				a.tpotMinS = tpotS
				a.tpotMaxS = tpotS
			} else {
				if tpotS < a.tpotMinS { a.tpotMinS = tpotS }
				if tpotS > a.tpotMaxS { a.tpotMaxS = tpotS }
			}
			a.tpotSumS += tpotS * tpotN
			a.tpotN += tpotN
		}
	}
	a.prevSnap = snap
	a.prevSet = true

	lat := ui.LatencyStats{}
	if a.ttftN > 0 {
		lat.TTFTMinMs = a.ttftMinS * 1000
		lat.TTFTAvgMs = a.ttftSumS / a.ttftN * 1000
		lat.TTFTMaxMs = a.ttftMaxS * 1000
	}
	if a.tpotN > 0 {
		lat.TPOTMinMs = a.tpotMinS * 1000
		lat.TPOTAvgMs = a.tpotSumS / a.tpotN * 1000
		lat.TPOTMaxMs = a.tpotMaxS * 1000
	}

	if gpuErr == nil && snap.GPUMemTotalMB > 0 {
		a.memHist.Push(snap.GPUUsedMB / snap.GPUMemTotalMB * 100)
	}
	a.kvHist.Push(snap.KVCacheUsagePct * 100)
	a.utilHist.Push(snap.GPUUtilPct)
	a.decHist.Push(delta.DecodeTokS)
	a.preHist.Push(delta.PrefillTokS)

	// Uptime from server process_start_time_seconds; fall back to app uptime if unavailable
	uptime := time.Since(a.startAt)
	if snap.StartTimeUnix > 0 {
		uptime = time.Since(time.Unix(int64(snap.StartTimeUnix), 0))
		if uptime < 0 {
			uptime = 0
		}
	}

	a.program.Send(ui.TickMsg{
		Backend: a.backend.Name(),
		GPUName: a.gpuName,
		Snap:    snap,
		Delta:   delta,
		Uptime:  uptime,
		DecHist: a.decHist.ValuesInto(a.decBuf),
		PreHist: a.preHist.ValuesInto(a.preBuf),
		MemHist: a.memHist.ValuesInto(a.memBuf),
		UtilHist: a.utilHist.ValuesInto(a.utilBuf),
		KVHist:   a.kvHist.ValuesInto(a.kvBuf),
		Latency:  lat,
	})
}
