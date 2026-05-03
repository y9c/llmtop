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

	// Prefix Cache
	PrefixCacheHits    float64
	PrefixCacheQueries float64

	// Latency histograms (from vLLM _sum / _count)
	TTFTTotalS   float64 // vllm:time_to_first_token_seconds_sum
	TTFTCount    float64 // vllm:time_to_first_token_seconds_count
	TPOTTotalS   float64 // vllm:request_time_per_output_token_seconds_sum
	TPOTCount    float64 // vllm:request_time_per_output_token_seconds_count

	// Session-wide latency tracking
	TTFTMinS   float64
	TTFTMaxS   float64
	TTFTAvgS   float64
	TPOTMinS   float64
	TPOTMaxS   float64
	TPOTAvgS   float64
	TTFTSamples float64
	TPOTSamples float64

	// Server start time (from process_start_time_seconds)
	StartTimeUnix float64
}

// Deltas holds per-second rates computed from two Snapshots.
type Deltas struct {
	DecodeTokS  float64
	PrefillTokS float64
	AcceptRate  float64
}

// GPU is a single GPU info row from nvidia-smi.
type GPU struct {
	Name     string
	UsedMB   float64
	TotalMB  float64
	UtilPct  float64
	TempC    float64
	PowerW   float64
	PowerMaxW float64
}

// IsEmpty returns true if this snapshot has no data yet.
func (s Snapshot) IsEmpty() bool {
	return s.GenTokensTotal == 0 && s.PromptTokensTotal == 0 && s.SpecDraftsTotal == 0
}

// ComputeDelta computes Deltas from two sequential snapshots.
func ComputeDelta(prev, cur Snapshot, dt float64) Deltas {
	var d Deltas
	if dt <= 0 || prev.IsEmpty() {
		return d
	}
	d.DecodeTokS = (cur.GenTokensTotal - prev.GenTokensTotal) / dt
	d.PrefillTokS = (cur.PromptTokensTotal - prev.PromptTokensTotal) / dt
	// AcceptRate = accepted tokens / draft tokens (0-1)
	if nd := cur.SpecDraftToksTotal - prev.SpecDraftToksTotal; nd > 0 {
		d.AcceptRate = (cur.SpecAcceptedTotal - prev.SpecAcceptedTotal) / nd
	}
	return d
}
