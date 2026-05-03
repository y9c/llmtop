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
			GPUTempC: 42, GPUPowerW: 82, GPUPowerMaxW: 350,
			RunningReqs: 1, WaitingReqs: 0,
			GenTokensTotal: 203300, PromptTokensTotal: 22290000,
			PromptCachedTotal: 19760000, PromptLocalTotal: 2530000,
			SpecDraftsTotal: 10900, SpecDraftToksTotal: 59300,
			SpecAcceptedTotal: 24100, PrefixCacheHits: 20360000,
			PrefixCacheQueries: 23450000,
			SpecAcceptedPos:    []float64{44800, 34000, 26400, 21400, 17700},
			TTFTTotalS:        100.77, TTFTCount: 30,
			TPOTTotalS: 0.69, TPOTCount: 30,
			StartTimeUnix: float64(time.Now().Unix() - 54600),
		},
		Delta:   metrics.Deltas{DecodeTokS: 48.0, PrefillTokS: 310.0, AcceptRate: 0.486},
		Uptime:  15*time.Hour + 14*time.Minute,
		DecHist: []float64{10, 20, 30, 25, 15, 5, 15, 25, 35, 42, 48, 45, 38, 28, 186},
		PreHist: []float64{100, 150, 200, 250, 300, 310, 305, 290, 280, 260, 250, 240, 230, 220, 210},
		MemHist: []float64{96, 97, 97, 96, 97, 97, 97, 96, 97, 97, 97, 96, 97, 97, 97},
		UtilHist: []float64{10, 30, 50, 70, 90, 95, 97, 96, 98, 97, 95, 90, 80, 60, 40},
		KVHist:  []float64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70},
		Width:   88,
	}
	fmt.Println(m.View())
}
