package backend

import "testing"

var llamaCPPSample = `llamacpp:prompt_tokens_total 1000
llamacpp:prompt_seconds_total 5.0
llamacpp:tokens_predicted_total 2000
llamacpp:tokens_predicted_seconds_total 10.0
llamacpp:n_decode_total 2000
llamacpp:n_tokens_max 2048
llamacpp:n_busy_slots_per_decode 1
llamacpp:prompt_tokens_seconds 200.0
llamacpp:predicted_tokens_seconds 200.0
llamacpp:requests_processing 2
llamacpp:requests_deferred 1
`

func TestLLamaCPPName(t *testing.T) {
	var b LLamaCPP
	if got := b.Name(); got != "llama.cpp" {
		t.Fatalf("Name(): want llama.cpp, got %s", got)
	}
}

func TestLLamaCPPDetect(t *testing.T) {
	var b LLamaCPP
	if !b.Detect("llamacpp:prompt_tokens_total 0") {
		t.Fatal("Detect(llamacpp:) should be true")
	}
	if b.Detect("other") {
		t.Fatal("Detect(other) should be false")
	}
}

func TestLLamaCPPParse(t *testing.T) {
	var b LLamaCPP
	s, err := b.Parse(llamaCPPSample)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if s.PromptTokensTotal != 1000 {
		t.Fatalf("PromptTokensTotal: want 1000, got %v", s.PromptTokensTotal)
	}
	if s.GenTokensTotal != 2000 {
		t.Fatalf("GenTokensTotal: want 2000, got %v", s.GenTokensTotal)
	}
	if s.RunningReqs != 2 {
		t.Fatalf("RunningReqs: want 2, got %v", s.RunningReqs)
	}
	if s.WaitingReqs != 1 {
		t.Fatalf("WaitingReqs: want 1, got %v", s.WaitingReqs)
	}
}
