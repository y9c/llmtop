package backend

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/changye/llmtop/internal/metrics"
)

type matchRule struct {
	key string // first unique word in the metric name for quick skip
	re  *regexp.Regexp
	set func(*metrics.Snapshot, float64)
}

var matchRules []matchRule

var reAcceptPos = regexp.MustCompile(
	`spec_decode_num_accepted_tokens_per_pos_total\{[^}]*position="(\d)"[^}]*\}\s+([\d.eE+-]+)`)

func init() {
	matchRules = []matchRule{
		{"num_requests_running", regexp.MustCompile(`num_requests_running\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.RunningReqs = v }},
		{"num_requests_waiting", regexp.MustCompile(`num_requests_waiting_by_reason\{[^}]*reason="capacity"[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.WaitingReqs = v }},
		{"kv_cache_usage_perc", regexp.MustCompile(`kv_cache_usage_perc\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.KVCacheUsagePct = v }},
		{"generation_tokens_total", regexp.MustCompile(`generation_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.GenTokensTotal = v }},
		{"prompt_tokens_total", regexp.MustCompile(`prompt_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptTokensTotal = v }},
		{"prompt_tokens_cached_total", regexp.MustCompile(`prompt_tokens_cached_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptCachedTotal = v }},
		{"prompt_tokens_by_source_total", regexp.MustCompile(`prompt_tokens_by_source_total\{[^}]*source="local_compute"[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptLocalTotal = v }},
		{"spec_decode_num_drafts_total", regexp.MustCompile(`spec_decode_num_drafts_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.SpecDraftsTotal = v }},
		{"spec_decode_num_draft_tokens_total", regexp.MustCompile(`spec_decode_num_draft_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.SpecDraftToksTotal = v }},
		{"spec_decode_num_accepted_tokens_total", regexp.MustCompile(`spec_decode_num_accepted_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.SpecAcceptedTotal = v }},
		{"prefix_cache_hits_total", regexp.MustCompile(`prefix_cache_hits_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PrefixCacheHits = v }},
		{"prefix_cache_queries_total", regexp.MustCompile(`prefix_cache_queries_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PrefixCacheQueries = v }},
		{"time_to_first_token", regexp.MustCompile(`time_to_first_token_seconds_sum\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TTFTTotalS = v }},
		{"time_to_first_token", regexp.MustCompile(`time_to_first_token_seconds_count\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TTFTCount = v }},
		{"request_time_per_output_token", regexp.MustCompile(`request_time_per_output_token_seconds_sum\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TPOTTotalS = v }},
		{"request_time_per_output_token", regexp.MustCompile(`request_time_per_output_token_seconds_count\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TPOTCount = v }},
	}
}

type VLLM struct{}

func (VLLM) Name() string { return "vLLM" }

func (VLLM) Detect(body string) bool { return strings.Contains(body, "vllm:") }

func (VLLM) Parse(body string) (metrics.Snapshot, error) {
	s := metrics.Snapshot{Timestamp: time.Now()}

	// Single pass per rule across the full body — ~18 regex calls instead of ~5400.
	for _, rule := range matchRules {
		// Quick skip: if the key word isn't in the body, skip the regex entirely.
		if !strings.Contains(body, rule.key) {
			continue
		}
		matches := rule.re.FindStringSubmatch(body)
		if len(matches) >= 2 {
			if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
				rule.set(&s, v)
			}
		}
	}

	// Speculative accept positions need special handling (dynamic key per position).
	if strings.Contains(body, "spec_decode_num_accepted_tokens_per_pos_total") {
		for _, m := range reAcceptPos.FindAllStringSubmatch(body, -1) {
			if v, err := strconv.ParseFloat(m[2], 64); err == nil {
				if idx, err := strconv.Atoi(m[1]); err == nil {
					for len(s.SpecAcceptedPos) <= idx {
						s.SpecAcceptedPos = append(s.SpecAcceptedPos, 0)
					}
					s.SpecAcceptedPos[idx] = v
				}
			}
		}
	}

	return s, nil
}
