package backend

import (
	"testing"
)

var sampleMetrics = `vllm:num_requests_running{engine="0",model_name="qwen3.6"} 1
vllm:num_requests_waiting_by_reason{engine="0",model_name="qwen3.6",reason="capacity"} 0
vllm:kv_cache_usage_perc{engine="0",model_name="qwen3.6"} 0.4565217391304348
vllm:generation_tokens_total{engine="0",model_name="qwen3.6"} 15000
vllm:prompt_tokens_total{engine="0",model_name="qwen3.6"} 2000000
vllm:prompt_tokens_cached_total{engine="0",model_name="qwen3.6"} 1900000
vllm:prompt_tokens_by_source_total{engine="0",model_name="qwen3.6",source="local_compute"} 100000
vllm:spec_decode_num_drafts_total{engine="0",model_name="qwen3.6"} 5000
vllm:spec_decode_num_draft_tokens_total{engine="0",model_name="qwen3.6"} 15000
vllm:spec_decode_num_accepted_tokens_total{engine="0",model_name="qwen3.6"} 10000
vllm:spec_decode_num_accepted_tokens_per_pos_total{engine="0",model_name="qwen3.6",position="0"} 5000
vllm:spec_decode_num_accepted_tokens_per_pos_total{engine="0",model_name="qwen3.6",position="1"} 3000
vllm:spec_decode_num_accepted_tokens_per_pos_total{engine="0",model_name="qwen3.6",position="2"} 2000
vllm:prefix_cache_hits_total{engine="0",model_name="qwen3.6"} 1900000
vllm:prefix_cache_queries_total{engine="0",model_name="qwen3.6"} 2000000
vllm:time_to_first_token_seconds_sum{engine="0",model_name="qwen3.6"} 45.678
vllm:time_to_first_token_seconds_count{engine="0",model_name="qwen3.6"} 100.0
vllm:request_time_per_output_token_seconds_sum{engine="0",model_name="qwen3.6"} 12.345
vllm:request_time_per_output_token_seconds_count{engine="0",model_name="qwen3.6"} 500.0
`

func TestVLLMName(t *testing.T) {
	var b VLLM
	if got := b.Name(); got != "vLLM" {
		t.Fatalf("Name(): want vLLM, got %s", got)
	}
}

func TestVLLMDetect(t *testing.T) {
	var b VLLM
	if !b.Detect("vllm:xxx") {
		t.Fatal("Detect(vllm:xxx) should be true")
	}
	if b.Detect("other") {
		t.Fatal("Detect(other) should be false")
	}
}

func TestVLLMParse(t *testing.T) {
	var b VLLM
	s, err := b.Parse(sampleMetrics)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if s.RunningReqs != 1 {
		t.Fatalf("RunningReqs: want 1, got %v", s.RunningReqs)
	}
	if s.GenTokensTotal != 15000 {
		t.Fatalf("GenTokensTotal: want 15000, got %v", s.GenTokensTotal)
	}
	if s.KVCacheUsagePct != 0.4565217391304348 {
		t.Fatalf("KVCacheUsagePct: want 0.4565217391304348, got %v", s.KVCacheUsagePct)
	}
	if s.SpecDraftsTotal != 5000 {
		t.Fatalf("SpecDraftsTotal: want 5000, got %v", s.SpecDraftsTotal)
	}
	if s.SpecAcceptedTotal != 10000 {
		t.Fatalf("SpecAcceptedTotal: want 10000, got %v", s.SpecAcceptedTotal)
	}
	if s.SpecAcceptedPos[0] != 5000 {
		t.Fatalf("SpecAcceptedPos[0]: want 5000, got %v", s.SpecAcceptedPos[0])
	}
	if s.SpecAcceptedPos[1] != 3000 {
		t.Fatalf("SpecAcceptedPos[1]: want 3000, got %v", s.SpecAcceptedPos[1])
	}
	if s.SpecAcceptedPos[2] != 2000 {
		t.Fatalf("SpecAcceptedPos[2]: want 2000, got %v", s.SpecAcceptedPos[2])
	}
	if s.PrefixCacheHits != 1900000 {
		t.Fatalf("PrefixCacheHits: want 1900000, got %v", s.PrefixCacheHits)
	}
	if s.TTFTTotalS != 45.678 {
		t.Fatalf("TTFTTotalS: want 45.678, got %v", s.TTFTTotalS)
	}
	if s.TTFTCount != 100 {
		t.Fatalf("TTFTCount: want 100, got %v", s.TTFTCount)
	}
	if s.TPOTTotalS != 12.345 {
		t.Fatalf("TPOTTotalS: want 12.345, got %v", s.TPOTTotalS)
	}
	if s.TPOTCount != 500 {
		t.Fatalf("TPOTCount: want 500, got %v", s.TPOTCount)
	}
}
