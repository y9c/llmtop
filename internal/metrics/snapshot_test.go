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

func TestComputeDeltaClampsAcceptRate(t *testing.T) {
	// Drafts advanced but accepted went down (counter reset) -> AcceptRate must be >= 0
	prev := Snapshot{SpecDraftToksTotal: 100, SpecAcceptedTotal: 60}
	cur := Snapshot{SpecDraftToksTotal: 150, SpecAcceptedTotal: 40}

	d := ComputeDelta(prev, cur, 1.0)
	if d.AcceptRate < 0 {
		t.Fatalf("AcceptRate should not be negative, got %v", d.AcceptRate)
	}
}
