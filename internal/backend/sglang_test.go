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
sglang:spec_accept_rate{tp_rank="0"} 0.7
sglang:time_to_first_token_seconds_sum{model_name="m",is_streaming="false"} 12.5
sglang:time_to_first_token_seconds_count{model_name="m",is_streaming="false"} 25
sglang:time_to_first_token_seconds_sum{model_name="m",is_streaming="true"} 10.5
sglang:time_to_first_token_seconds_count{model_name="m",is_streaming="true"} 15
sglang:inter_token_latency_seconds_sum{model_name="m"} 30.0
sglang:inter_token_latency_seconds_count{model_name="m"} 100
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

func TestSGLangDetectUnprefixedOldNames(t *testing.T) {
	var b SGLang
	body := "num_running_requests 2\nnum_waiting_requests 1\ntoken_usage 0.5\ncache_hit_rate 0.3\n"
	if !b.Detect(body) {
		t.Fatal("Detect(unprefixed pre-0.5 body) should be true")
	}
}

func TestSGLangDetectRejectsVLLM(t *testing.T) {
	var b SGLang
	body := "vllm:num_requests_running 1\nvllm:num_requests_waiting 0\nvllm:kv_cache_usage_perc 0.5\nvllm:prompt_tokens_total 100\n"
	if b.Detect(body) {
		t.Fatal("genuine vLLM body must not detect as SGLang")
	}
}

func TestSGLangParseUnprefixedOldNames(t *testing.T) {
	var b SGLang
	body := `num_running_requests 3
num_waiting_requests 2
token_usage 0.42
cache_hit_rate 0.31
prompt_tokens_total 5000
generation_tokens_total 15000
time_to_first_token_seconds_sum 12.5
time_to_first_token_seconds_count 25
`
	s, err := b.Parse(body)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if s.RunningReqs != 3 {
		t.Fatalf("RunningReqs: want 3, got %v", s.RunningReqs)
	}
	if s.WaitingReqs != 2 {
		t.Fatalf("WaitingReqs: want 2, got %v", s.WaitingReqs)
	}
	if s.KVCacheUsagePct != 0.42 {
		t.Fatalf("KVCacheUsagePct: want 0.42, got %v", s.KVCacheUsagePct)
	}
	if s.PrefixCacheHits != 0.31 {
		t.Fatalf("PrefixCacheHits: want 0.31, got %v", s.PrefixCacheHits)
	}
	if s.PromptTokensTotal != 5000 {
		t.Fatalf("PromptTokensTotal: want 5000, got %v", s.PromptTokensTotal)
	}
	if s.GenTokensTotal != 15000 {
		t.Fatalf("GenTokensTotal: want 15000, got %v", s.GenTokensTotal)
	}
	if s.TTFTCount != 25 || s.TTFTTotalS != 12.5 {
		t.Fatalf("TTFT: want count=25 sum=12.5, got %v/%v", s.TTFTCount, s.TTFTTotalS)
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
	// no reconstructed cumulative totals: the server exports no such counters
	if s.SpecDraftToksTotal != 0 || s.SpecAcceptedTotal != 0 {
		t.Fatalf("spec totals must not be reconstructed, got drafted=%v accepted=%v", s.SpecDraftToksTotal, s.SpecAcceptedTotal)
	}
	if s.SpecAcceptLen != 2.5 {
		t.Fatalf("SpecAcceptLen: want 2.5, got %v", s.SpecAcceptLen)
	}
	if s.SpecAcceptRate != 0.7 {
		t.Fatalf("SpecAcceptRate: want 0.7, got %v", s.SpecAcceptRate)
	}
	if s.TTFTCount != 40 || s.TTFTTotalS != 23 {
		t.Fatalf("TTFT: want count=40 sum=23 (both streaming series summed), got %v/%v", s.TTFTCount, s.TTFTTotalS)
	}
	if s.TPOTCount != 100 || s.TPOTTotalS != 30.0 {
		t.Fatalf("TPOT: want count=100 sum=30, got %v/%v", s.TPOTCount, s.TPOTTotalS)
	}
}
