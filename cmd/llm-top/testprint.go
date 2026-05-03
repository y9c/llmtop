//go:build ignore

package main

import (
	"fmt"
	"time"
	"github.com/y9c/llmtop/internal/metrics"
	"github.com/y9c/llmtop/internal/ui"
)

func main() {
	m := ui.Model{
		Backend: "vllm", GPUName: "L40S",
		Snap: metrics.Snapshot{
			GPUName: "NVIDIA L40S", GPUUsedMB: 43600, GPUMemTotalMB: 45000,
			GPUUtilPct: 97.0, KVCacheUsagePct: 0.093,
			RunningReqs: 1, WaitingReqs: 0,
			GenTokensTotal: 35500, PromptTokensTotal: 513500,
			PromptCachedTotal: 358800, PromptLocalTotal: 154700,
			SpecDraftsTotal: 10900, SpecDraftToksTotal: 41000,
			SpecAcceptedTotal: 24100, PrefixCacheHits: 300000,
			PrefixCacheQueries: 500000,
			SpecAcceptedPos: []float64{8300, 6000, 4400, 3400, 2600},
		},
		Delta: metrics.Deltas{DecodeTokS: 48.0, PrefillTokS: 0.0, AcceptRate: 0.486},
		Uptime: 20 * time.Second,
		DecHist: []float64{10, 20, 30, 25, 15, 5, 15, 25, 35, 42, 48, 45, 38, 28, 18},
		MemHist: []float64{96, 97, 97, 96, 97, 97, 97, 96, 97, 97, 97, 96, 97, 97, 97},
		UtilHist: []float64{10, 30, 50, 70, 90, 95, 97, 96, 98, 97, 95, 90, 80, 60, 40},
		KVHist:  []float64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70},
		Width:   88,
	}
	fmt.Println(m.View())
}
