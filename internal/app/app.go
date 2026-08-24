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

	// Last valid instantaneous TTFT/TPOT (held across ticks that have no new data)
	lastInstTTFT float64
	lastInstTPOT float64

	// Cumulative average (only ticks with new activity)
	decCumSum   float64
	decCumCount float64
	preCumSum   float64
	preCumCount float64
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
	gpuCtx, gpuCancel := context.WithTimeout(ctx, 3*time.Second)
	defer gpuCancel()

	gpuList, gpuErr := a.gpu.Fetch(gpuCtx)
	if gpuErr == nil && a.gpuName == "" && len(gpuList) > 0 {
		a.gpuName = gpuList[0].Name
	}

	httpCtx, httpCancel := context.WithTimeout(ctx, 3*time.Second)
	defer httpCancel()

	body, httpErr := a.fetcher.Fetch(httpCtx, a.cfg.MetricsURL())

	if a.cfg.Backend != "auto" && a.backend == nil {
		a.backend = backend.ByName(a.cfg.Backend)
	}
	if a.backend == nil && httpErr == nil {
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
		snap.GPUs = gpuList
		var avg metrics.GPU
		for _, g := range gpuList {
			avg.UsedMB += g.UsedMB
			avg.TotalMB += g.TotalMB
			avg.UtilPct += g.UtilPct
			avg.TempC += g.TempC
			avg.PowerW += g.PowerW
			avg.PowerMaxW += g.PowerMaxW
		}
		n := float64(len(gpuList))
		avg.Name = gpuList[0].Name
		avg.UsedMB /= n
		avg.TotalMB /= n
		avg.UtilPct /= n
		avg.TempC /= n
		avg.PowerW /= n
		avg.PowerMaxW /= n
		snap.GPUUsedMB = avg.UsedMB
		snap.GPUMemTotalMB = avg.TotalMB
		snap.GPUUtilPct = avg.UtilPct
		snap.GPUTempC = avg.TempC
		snap.GPUPowerW = avg.PowerW
		snap.GPUPowerMaxW = avg.PowerMaxW
		snap.GPUName = avg.Name
	}

	var delta metrics.Deltas
	if a.prevSet {
		delta = metrics.ComputeDelta(a.prevSnap, snap, a.cfg.Rate.Seconds())

		// Track per-sample TTFT/TPOT from histogram deltas
		if ttftN := snap.TTFTCount - a.prevSnap.TTFTCount; ttftN > 0 {
			a.lastInstTTFT = (snap.TTFTTotalS - a.prevSnap.TTFTTotalS) / ttftN * 1000
		}
		if tpotN := snap.TPOTCount - a.prevSnap.TPOTCount; tpotN > 0 {
			a.lastInstTPOT = (snap.TPOTTotalS - a.prevSnap.TPOTTotalS) / tpotN * 1000
		}
	}
	a.prevSnap = snap
	if !snap.IsEmpty() {
		a.prevSet = true
	}

	// Cumulative averages only (don't carry forward instantaneous values —
	// if no new tokens were generated, show 0, consistent with run=0).
	if delta.DecodeTokS > 0 {
		a.decCumSum += delta.DecodeTokS
		a.decCumCount++
	}
	if delta.PrefillTokS > 0 {
		a.preCumSum += delta.PrefillTokS
		a.preCumCount++
	}
	delta.TTFTMs = a.lastInstTTFT
	delta.TPOTMs = a.lastInstTPOT

	delta.DecCumAvg = 0
	if a.decCumCount > 0 {
		delta.DecCumAvg = a.decCumSum / a.decCumCount
	}
	delta.PreCumAvg = 0
	if a.preCumCount > 0 {
		delta.PreCumAvg = a.preCumSum / a.preCumCount
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
	})
}
