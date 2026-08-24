package backend

import "testing"

var sglangSample = `sglang:prompt_tokens_total{model_name="m",is_streaming="false"} 5000
sglang:prompt_tokens_total{model_name="m",is_streaming="true"} 7000
sglang:generation_tokens_total{model_name="m",is_streaming="false"} 15000
sglang:generation_tokens_total{model_name="m",is_streaming="true"} 3000
sglang:num_running_reqs{tp_rank="0"} 3
sglang:num_queue_reqs{tp_rank="0"} 2
sglang:token_usage{tp_rank="0"} 0.85
sglang:cache_hit_rate{tp_rank="0"} 0.31
sglang:spec_verify_calls_total{engine_type="unified"} 40
sglang:spec_accept_length{tp_rank="0"} 2.5
sglang:spec_num_draft_tokens{tp_rank="0"} 10
sglang:time_to_first_token_seconds_sum{model_name="m"} 12.5
sglang:time_to_first_token_seconds_count{model_name="m"} 25
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

	if s.PromptTokensTotal != 12000 {
		t.Fatalf("PromptTokensTotal: want 12000 (both streaming series summed), got %v", s.PromptTokensTotal)
	}
	if s.GenTokensTotal != 18000 {
		t.Fatalf("GenTokensTotal: want 18000 (both streaming series summed), got %v", s.GenTokensTotal)
	}
	if s.RunningReqs != 3 {
		t.Fatalf("RunningReqs: want 3, got %v", s.RunningReqs)
	}
	if s.WaitingReqs != 2 {
		t.Fatalf("WaitingReqs: want 2, got %v", s.WaitingReqs)
	}
	if s.KVCacheUsagePct != 0.85 {
		t.Fatalf("KVCacheUsagePct: want 0.85 (fraction), got %v", s.KVCacheUsagePct)
	}
	if s.SpecDraftsTotal != 40 {
		t.Fatalf("SpecDraftsTotal: want 40, got %v", s.SpecDraftsTotal)
	}
	// reconstructed cumulative totals: verify_calls * draft_cap / * accept_len
	if s.SpecDraftToksTotal != 400 {
		t.Fatalf("SpecDraftToksTotal: want 400, got %v", s.SpecDraftToksTotal)
	}
	if s.SpecAcceptedTotal != 100 {
		t.Fatalf("SpecAcceptedTotal: want 100, got %v", s.SpecAcceptedTotal)
	}
	if s.TTFTCount != 25 || s.TTFTTotalS != 12.5 {
		t.Fatalf("TTFT: want count=25 sum=12.5, got %v/%v", s.TTFTCount, s.TTFTTotalS)
	}
}
