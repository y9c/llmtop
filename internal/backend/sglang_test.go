package backend

import "testing"

var sglangSample = `sglang:prompt_tokens_total 5000
sglang:generation_tokens_total 15000
sglang:num_running_requests 3
sglang:num_waiting_requests 2
sglang:prefix_cache_hit_rate 0.85
`

func TestSGLangName(t *testing.T) {
	var b SGLang
	if got := b.Name(); got != "SGLang" {
		t.Fatalf("Name(): want SGLang, got %s", got)
	}
}

func TestSGLangDetect(t *testing.T) {
	var b SGLang
	if !b.Detect("sglang:prompt_tokens_total 0") {
		t.Fatal("Detect(sglang:) should be true")
	}
	if b.Detect("other") {
		t.Fatal("Detect(other) should be false")
	}
}

func TestSGLangParse(t *testing.T) {
	var b SGLang
	s, err := b.Parse(sglangSample)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if s.PromptTokensTotal != 5000 {
		t.Fatalf("PromptTokensTotal: want 5000, got %v", s.PromptTokensTotal)
	}
	if s.GenTokensTotal != 15000 {
		t.Fatalf("GenTokensTotal: want 15000, got %v", s.GenTokensTotal)
	}
	if s.RunningReqs != 3 {
		t.Fatalf("RunningReqs: want 3, got %v", s.RunningReqs)
	}
	if s.WaitingReqs != 2 {
		t.Fatalf("WaitingReqs: want 2, got %v", s.WaitingReqs)
	}
	if s.KVCacheUsagePct != 85.0 {
		t.Fatalf("KVCacheUsagePct: want 85.0, got %v", s.KVCacheUsagePct)
	}
}
