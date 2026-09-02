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

// rateAvgWindow is the rolling window over which bursty prefill/decode
// rates are smoothed for display.
const rateAvgWindow = 5 * time.Second

type App struct {
	cfg     *config.Config
	fetcher *fetcher.Fetcher
	gpu     collector.GPUCollector
	program *tea.Program
	model   *ui.Model

	memHist  *metrics.History
	kvHist   *metrics.History
	utilHist *metrics.History
	decHist  *metrics.History
	preHist  *metrics.History

	// Rolling-rate windows used to smooth bursty prefill/decode rates
	decRateHist *metrics.History
	preRateHist *metrics.History

	prevSnap    metrics.Snapshot
	prevSet     bool
	backend     backend.Backend
	startAt     time.Time
	resolvedURL string
	gpuName     string

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
		cfg:         cfg,
		fetcher:     f,
		gpu:         gpu,
		model:       m,
		memHist:     metrics.NewHistory(),
		kvHist:      metrics.NewHistory(),
		utilHist:    metrics.NewHistory(),
		decHist:     metrics.NewHistory(),
		preHist:     metrics.NewHistory(),
		decRateHist: metrics.NewHistory(),
		preRateHist: metrics.NewHistory(),
		startAt:     time.Now(),
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

	url := a.cfg.MetricsURL()
	if a.resolvedURL != "" {
		url = a.resolvedURL
	}
	body, httpErr := a.fetcher.Fetch(httpCtx, url)

	if a.cfg.Backend != "auto" && a.backend == nil {
		a.backend = backend.ByName(a.cfg.Backend)
	}
	if httpErr != nil && a.resolvedURL == "" && a.cfg.Probeable() {
		if pbody, ok := a.probeAlternates(httpCtx); ok {
			body, httpErr = pbody, nil
		}
	}
	if a.backend == nil && httpErr == nil {
		a.backend = backend.Detect(body)
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
		if len(gpuList) > 0 {
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
	}

	var delta metrics.Deltas
	if a.prevSet {
		// Use the actual wall time between samples so rates stay correct when
		// a scrape stalls or two ticks land back-to-back.
		dt := snap.Timestamp.Sub(a.prevSnap.Timestamp).Seconds()
		delta = metrics.ComputeDelta(a.prevSnap, snap, dt)

		if !snap.IsEmpty() {
			// Track per-sample TTFT/TPOT from histogram deltas
			if ttftN := snap.TTFTCount - a.prevSnap.TTFTCount; ttftN > 0 {
				a.lastInstTTFT = (snap.TTFTTotalS - a.prevSnap.TTFTTotalS) / ttftN * 1000
			} else if ttftN < 0 {
				a.lastInstTTFT = 0 // server histograms reset
			}
			if tpotN := snap.TPOTCount - a.prevSnap.TPOTCount; tpotN > 0 {
				a.lastInstTPOT = (snap.TPOTTotalS - a.prevSnap.TPOTTotalS) / tpotN * 1000
			} else if tpotN < 0 {
				a.lastInstTPOT = 0
			}
		}
	}
	// Only advance the delta baseline with real data; a failed scrape must not
	// be treated as "counters went to zero" (that would inflate the next sample).
	if !snap.IsEmpty() {
		a.prevSnap = snap
		a.prevSet = true
	}

	// Cumulative averages use the raw instantaneous samples.
	if delta.DecodeTokS > 0 {
		a.decCumSum += delta.DecodeTokS
		a.decCumCount++
	}
	if delta.PrefillTokS > 0 {
		a.preCumSum += delta.PrefillTokS
		a.preCumCount++
	}

	// Smooth bursty rates over a rolling window so prefill/decode bursts
	// don't snap to 0 between requests.
	window := int(rateAvgWindow / a.cfg.Rate)
	if window < 1 {
		window = 1
	}
	a.decRateHist.Push(delta.DecodeTokS)
	a.preRateHist.Push(delta.PrefillTokS)
	delta.DecodeTokS = a.decRateHist.RecentAvg(window)
	delta.PrefillTokS = a.preRateHist.RecentAvg(window)

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

	backendName := "—"
	if a.backend != nil {
		backendName = a.backend.Name()
	}

	a.program.Send(ui.TickMsg{
		Backend:  backendName,
		GPUName:  a.gpuName,
		Snap:     snap,
		Delta:    delta,
		Uptime:   uptime,
		DecHist:  a.decHist.ValuesInto(),
		PreHist:  a.preHist.ValuesInto(),
		MemHist:  a.memHist.ValuesInto(),
		UtilHist: a.utilHist.ValuesInto(),
		KVHist:   a.kvHist.ValuesInto(),
	})
}

// probeAlternates scans the candidate ports for a live metrics endpoint and
// records the winner in resolvedURL. It returns the winning body so the
// caller can parse it in the same tick.
func (a *App) probeAlternates(ctx context.Context) (string, bool) {
	for _, u := range a.cfg.ProbeURLs() {
		body, err := a.fetcher.Probe(ctx, u)
		if err != nil {
			continue
		}
		b, ok := backend.KnownDetect(body)
		if !ok {
			continue
		}
		a.resolvedURL = u
		if a.backend == nil {
			a.backend = b
		}
		return body, true
	}
	return "", false
}
