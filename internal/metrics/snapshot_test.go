package metrics

import "testing"

func TestComputeDeltaClampsNegativeRates(t *testing.T) {
	prev := Snapshot{GenTokensTotal: 1000, PromptTokensTotal: 5000}
	cur := Snapshot{GenTokensTotal: 500, PromptTokensTotal: 2000} // counter reset downward

	d := ComputeDelta(prev, cur, 1.0)
	if d.DecodeTokS != 0 {
		t.Fatalf("DecodeTokS should clamp to 0 on counter reset, got %v", d.DecodeTokS)
	}
	if d.PrefillTokS != 0 {
		t.Fatalf("PrefillTokS should clamp to 0 on counter reset, got %v", d.PrefillTokS)
	}
}

func TestComputeDeltaPositiveRates(t *testing.T) {
	prev := Snapshot{GenTokensTotal: 1000, PromptTokensTotal: 5000}
	cur := Snapshot{GenTokensTotal: 1200, PromptTokensTotal: 5300}

	d := ComputeDelta(prev, cur, 1.0)
	if d.DecodeTokS != 200 {
		t.Fatalf("DecodeTokS: want 200, got %v", d.DecodeTokS)
	}
	if d.PrefillTokS != 300 {
		t.Fatalf("PrefillTokS: want 300, got %v", d.PrefillTokS)
	}
}

func TestComputeDeltaPrefersLiveGauge(t *testing.T) {
	// Counter delta would say 3000 tok/s (completion-impulsed), but the
	// server's live windowed gauge (gen_throughput) is the honest rate.
	prev := Snapshot{GenTokensTotal: 1000}
	cur := Snapshot{GenTokensTotal: 4000, GenTokS: 892.5}

	d := ComputeDelta(prev, cur, 1.0)
	if d.DecodeTokS != 892.5 {
		t.Fatalf("DecodeTokS: want 892.5 (live gauge overrides counter delta), got %v", d.DecodeTokS)
	}
}

func TestComputeDeltaKeepsCounterWhenNoGauge(t *testing.T) {
	// vLLM path: no live gauge, so keep the counter delta.
	prev := Snapshot{GenTokensTotal: 1000}
	cur := Snapshot{GenTokensTotal: 1200}

	d := ComputeDelta(prev, cur, 1.0)
	if d.DecodeTokS != 200 {
		t.Fatalf("DecodeTokS: want 200 (counter delta when no gauge), got %v", d.DecodeTokS)
	}
}

func TestComputeDeltaPrefersSmoothPrefill(t *testing.T) {
	// prefill_effective_tokens_total incs per batch (smooth); use its delta
	// instead of prompt_tokens_total which jumps only at request completion.
	prev := Snapshot{PrefillEffectiveTotal: 40000, PromptTokensTotal: 50000}
	cur := Snapshot{PrefillEffectiveTotal: 40300, PromptTokensTotal: 90000}

	d := ComputeDelta(prev, cur, 1.0)
	if d.PrefillTokS != 300 { // 300 from smooth prefill, not 40000 from prompt jump
		t.Fatalf("PrefillTokS: want 300 (smooth prefill delta), got %v", d.PrefillTokS)
	}
}

func TestComputeDeltaClampsAcceptRate(t *testing.T) {
	// Drafts advanced but accepted went down (counter reset) -> AcceptRate must be >= 0
	prev := Snapshot{SpecDraftToksTotal: 100, SpecAcceptedTotal: 60}
	cur := Snapshot{SpecDraftToksTotal: 150, SpecAcceptedTotal: 40}

	d := ComputeDelta(prev, cur, 1.0)
	if d.AcceptRate < 0 {
		t.Fatalf("AcceptRate should not be negative, got %v", d.AcceptRate)
	}
}
