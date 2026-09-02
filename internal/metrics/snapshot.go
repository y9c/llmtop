package metrics

import "time"

// Snapshot holds one sampling point of all metrics.
type Snapshot struct {
	Timestamp time.Time

	// GPU
	GPUName         string
	GPUUsedMB       float64
	GPUMemTotalMB   float64
	GPUUtilPct      float64
	GPUTempC        float64
	GPUPowerW       float64
	GPUPowerMaxW    float64
	KVCacheUsagePct float64

	GPUs []GPU

	// Requests
	RunningReqs float64
	WaitingReqs float64

	// Tokens
	GenTokensTotal    float64
	PromptTokensTotal float64
	PromptCachedTotal float64
	PromptLocalTotal  float64

	// Speculative Decoding (MTP / Eagle / DFlash)
	SpecDraftsTotal    float64
	SpecDraftToksTotal float64
	SpecAcceptedTotal  float64
	SpecAcceptedPos    []float64
	SpecAcceptLen      float64 // server-side mean accepted tokens per verify pass (SGLang gauge)
	SpecAcceptRate     float64 // server-side acceptance rate 0-1 (SGLang gauge)

	// Prefix Cache
	PrefixCacheHits    float64
	PrefixCacheQueries float64

	// Latency histograms (from vLLM _sum / _count)
	TTFTTotalS float64 // vllm:time_to_first_token_seconds_sum
	TTFTCount  float64 // vllm:time_to_first_token_seconds_count
	TPOTTotalS float64 // vllm:request_time_per_output_token_seconds_sum
	TPOTCount  float64 // vllm:request_time_per_output_token_seconds_count

	// Server start time (from process_start_time_seconds)
	StartTimeUnix float64
}

// Deltas holds per-second rates computed from two Snapshots.
type Deltas struct {
	DecodeTokS  float64 // rolling-window decode tok/s (0 when idle)
	PrefillTokS float64 // rolling-window prefill tok/s (0 when idle)
	DecCumAvg   float64 // cumulative average decode tok/s (only active ticks)
	PreCumAvg   float64 // cumulative average prefill tok/s (only active ticks)
	TTFTMs      float64 // last-sample TTFT ms (held across ticks with no new data)
	TPOTMs      float64 // last-sample TPOT ms (held across ticks with no new data)
	AcceptRate  float64
}

func (s Snapshot) GPUCount() int { return len(s.GPUs) }

// GPU is a single GPU info row from nvidia-smi.
type GPU struct {
	Name      string
	UsedMB    float64
	TotalMB   float64
	UtilPct   float64
	TempC     float64
	PowerW    float64
	PowerMaxW float64
}

// IsEmpty returns true if this snapshot has no data yet.
func (s Snapshot) IsEmpty() bool {
	return s.GenTokensTotal == 0 && s.SpecDraftsTotal == 0 && s.PromptTokensTotal == 0
}

// ComputeDelta computes Deltas from two sequential snapshots.
func ComputeDelta(prev, cur Snapshot, dt float64) Deltas {
	var d Deltas
	if dt <= 0 || prev.IsEmpty() {
		return d
	}
	d.DecodeTokS = (cur.GenTokensTotal - prev.GenTokensTotal) / dt
	d.PrefillTokS = (cur.PromptTokensTotal - prev.PromptTokensTotal) / dt
	// Token rates are monotonic; a negative delta means the server's counters
	// reset (restart/relabel), so surface it as 0 rather than a negative rate.
	if d.DecodeTokS < 0 {
		d.DecodeTokS = 0
	}
	if d.PrefillTokS < 0 {
		d.PrefillTokS = 0
	}
	// AcceptRate = accepted tokens / draft tokens (0-1)
	if nd := cur.SpecDraftToksTotal - prev.SpecDraftToksTotal; nd > 0 {
		d.AcceptRate = (cur.SpecAcceptedTotal - prev.SpecAcceptedTotal) / nd
	}
	if d.AcceptRate < 0 {
		d.AcceptRate = 0
	}
	if d.AcceptRate > 1 {
		d.AcceptRate = 1
	}
	return d
}
