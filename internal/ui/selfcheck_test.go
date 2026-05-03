package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/y9c/llmtop/internal/metrics"
)

func TestSelfCheckAlignmentAllWidths(t *testing.T) {
	widths := []int{80, 82, 84, 86, 88, 90, 100}
	for _, w := range widths {
		t.Run(fmt.Sprintf("w=%d", w), func(t *testing.T) {
			m := buildTestModel(w)
			output := m.buildView()
			lines := splitLinesS(output)
			checkPipes(t, lines, w)
		})
	}
}

func TestSelfCheckAlignmentZeroData(t *testing.T) {
	m := buildZeroModel(88)
	output := m.buildView()
	lines := splitLinesS(output)
	checkPipes(t, lines, 88)
}

func checkPipes(t *testing.T, lines []string, cw int) {
	boxType := ""
	// We track positions of the first and second levels of pipes
	// Level 1: '│' at the very left edge
	// Level 2: '│' at the column separator (two-col) or right edge
	var firstPipe, secondPipe []int

	for _, raw := range lines {
		plain := raw
		// Replace all ANSI sequences with "X" for width-measurement purposes
		plain = ansiToX(plain)

		if strings.HasPrefix(plain, "┌") {
			if strings.Contains(plain, "┬") {
				boxType = "twoCol"
			} else {
				boxType = "single"
			}
			firstPipe = nil
			secondPipe = nil
			continue
		}
		if strings.HasPrefix(plain, "└") || strings.HasPrefix(plain, "╘") {
			boxType = ""
			continue
		}
		if boxType == "" {
			continue
		}

		pp := pipePos(plain)
		if len(pp) < 2 {
			continue
		}

		if firstPipe == nil {
			firstPipe = []int{pp[0]}
			if len(pp) >= 2 {
				secondPipe = []int{pp[1]}
			}
			continue
		}

		if pp[0] != firstPipe[0] {
			t.Errorf("[%s] left│ mismatch at %d vs %d: %q", boxType, pp[0], firstPipe[0], truncS(plain, 60))
		}
		if boxType == "twoCol" && len(pp) >= 3 && secondPipe != nil {
			if pp[1] != secondPipe[0] {
				t.Errorf("[%s] mid│(%s) mismatch %d vs %d: %q", boxType, "w=", pp[1], secondPipe[0], "")
			}
		}
	}
}

// Replace ANSI escape sequences with 'X' markers for consistent width measurement
func ansiToX(s string) string {
	var out strings.Builder
	in := false
	for _, r := range s {
		if r == '\x1b' {
			in = true
		}
		if in {
			if r == 'm' {
				in = false
				out.WriteRune('█') // 1 char replacement
			}
			continue
		}
		out.WriteRune(r)
	}
	return out.String()
}

func pipePos(s string) []int {
	var p []int
	for i, r := range s {
		if r == '│' || r == '┤' || r == '┼' || r == '┴' {
			p = append(p, i)
		}
	}
	return p
}

func splitLinesS(s string) []string {
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

func truncS(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n]) + "..."
}

func buildTestModel(w int) Model {
	return Model{
		Backend: "vllm",
		GPUName: "L40S",
		Snap: metrics.Snapshot{
			GPUName: "NVIDIA L40S", GPUUsedMB: 43600, GPUMemTotalMB: 45000,
			GPUUtilPct: 97.0, KVCacheUsagePct: 0.093,
			RunningReqs: 1, WaitingReqs: 0,
			GenTokensTotal: 35500, PromptTokensTotal: 513500,
			PromptCachedTotal: 358800, PromptLocalTotal: 154700,
			SpecDraftsTotal: 10900, SpecDraftToksTotal: 41000,
			SpecAcceptedTotal: 24100, PrefixCacheHits: 300000,
			PrefixCacheQueries: 500000,
			SpecAcceptedPos:   []float64{8300, 6000, 4400, 3400, 2600},
		},
		Delta:   metrics.Deltas{DecodeTokS: 48.0, PrefillTokS: 0.0, AcceptRate: 0.486},
		Uptime:  20 * time.Second,
		DecHist: []float64{10, 20, 30, 25, 15, 5, 15, 25, 35, 42, 48, 45, 38, 28, 18},
		MemHist: []float64{96, 97, 97, 96, 97, 97, 97, 96, 97, 97, 97, 96, 97, 97, 97},
		UtilHist: []float64{10, 30, 50, 70, 90, 95, 97, 96, 98, 97, 95, 90, 80, 60, 40},
		KVHist:  []float64{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70},
		Width:   w,
	}
}

func BenchmarkBuildView(b *testing.B) {
	m := buildTestModel(88)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.buildView()
	}
}

func buildZeroModel(w int) Model {
	return Model{
		Backend: "vllm", GPUName: "L40S",
		Snap: metrics.Snapshot{
			GPUName: "NVIDIA L40S", GPUUsedMB: 43600, GPUMemTotalMB: 45000,
			GPUUtilPct: 0, KVCacheUsagePct: 0,
			SpecAcceptedPos:  []float64{11200, 8100, 6000, 4600, 3600},
			SpecDraftsTotal: 14900, SpecDraftToksTotal: 56000,
			SpecAcceptedTotal: 33900, GenTokensTotal: 48300,
			PromptTokensTotal: 514900, PromptCachedTotal: 358800,
			PromptLocalTotal: 156200,
		},
		Delta:   metrics.Deltas{},
		Uptime:  8 * time.Second,
		DecHist: []float64{0, 10, 20, 30, 22, 28, 18},
		MemHist: []float64{96, 96, 96, 96, 96, 96, 96},
		UtilHist: []float64{0, 0, 0, 0, 0, 0, 0},
		KVHist:  []float64{0, 0, 0, 0, 0, 0, 0},
		Width:   w,
	}
}
