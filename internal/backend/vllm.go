package backend

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/y9c/llmtop/internal/metrics"
)

type matchRule struct {
	key string // first unique word in the metric name for quick skip
	re  *regexp.Regexp
	set func(*metrics.Snapshot, float64)
	sum bool // sum all labeled series (counters) instead of taking the first
}

var matchRules []matchRule

var reAcceptPos = regexp.MustCompile(
	`spec_decode_num_accepted_tokens_per_pos_total\{[^}]*position="(\d+)"[^}]*\}\s+([\d.eE+-]+)`)

func init() {
	// vLLM 0.20.1 adds a "vllm:" prefix, older versions don't.
	// Make prefix optional: (?:vllm:)? to match both formats.
	matchRules = []matchRule{
		{"num_requests_running", regexp.MustCompile(`(?:vllm:)?num_requests_running\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.RunningReqs = v }, true},
		{"num_requests_waiting", regexp.MustCompile(`(?:vllm:)?num_requests_waiting\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.WaitingReqs = v }, true},
		{"kv_cache_usage_perc", regexp.MustCompile(`(?:vllm:)?kv_cache_usage_perc\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.KVCacheUsagePct = v }, false},
		{"generation_tokens_total", regexp.MustCompile(`(?:vllm:)?generation_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.GenTokensTotal = v }, true},
		{"prompt_tokens_total", regexp.MustCompile(`(?:vllm:)?prompt_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptTokensTotal = v }, true},
		{"prompt_tokens_cached_total", regexp.MustCompile(`(?:vllm:)?prompt_tokens_cached_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptCachedTotal = v }, true},
		{"prompt_tokens_by_source_total", regexp.MustCompile(`(?:vllm:)?prompt_tokens_by_source_total\{[^}]*source="local_compute"[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PromptLocalTotal = v }, true},
		{"spec_decode_num_drafts_total", regexp.MustCompile(`(?:vllm:)?spec_decode_num_drafts_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.SpecDraftsTotal = v }, true},
		{"spec_decode_num_draft_tokens_total", regexp.MustCompile(`(?:vllm:)?spec_decode_num_draft_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.SpecDraftToksTotal = v }, true},
		{"spec_decode_num_accepted_tokens_total", regexp.MustCompile(`(?:vllm:)?spec_decode_num_accepted_tokens_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.SpecAcceptedTotal = v }, true},
		{"prefix_cache_hits_total", regexp.MustCompile(`(?:vllm:)?prefix_cache_hits_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PrefixCacheHits = v }, true},
		{"prefix_cache_queries_total", regexp.MustCompile(`(?:vllm:)?prefix_cache_queries_total\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.PrefixCacheQueries = v }, true},
		{"time_to_first_token", regexp.MustCompile(`(?:vllm:)?time_to_first_token_seconds_sum\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TTFTTotalS = v }, true},
		{"time_to_first_token", regexp.MustCompile(`(?:vllm:)?time_to_first_token_seconds_count\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TTFTCount = v }, true},
		{"request_time_per_output_token", regexp.MustCompile(`(?:vllm:)?request_time_per_output_token_seconds_sum\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TPOTTotalS = v }, true},
		{"request_time_per_output_token", regexp.MustCompile(`(?:vllm:)?request_time_per_output_token_seconds_count\{[^}]*\}\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.TPOTCount = v }, true},
		{"process_start_time_seconds", regexp.MustCompile(`process_start_time_seconds\s+([\d.eE+-]+)`), func(s *metrics.Snapshot, v float64) { s.StartTimeUnix = v }, false},
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
		if rule.sum {
			// Counter metrics can span multiple labeled series (one per engine /
			// model); total them so nothing is under-counted.
			var total float64
			for _, m := range rule.re.FindAllStringSubmatch(body, -1) {
				if v, err := strconv.ParseFloat(m[1], 64); err == nil {
					total += v
				}
			}
			rule.set(&s, total)
			continue
		}
		matches := rule.re.FindStringSubmatch(body)
		if len(matches) >= 2 {
			if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
				rule.set(&s, v)
			}
		}
	}

	// vLLM also exports num_requests_waiting_by_reason (one labeled series per
	// reason). If plain num_requests_waiting wasn't present, total the reasons
	// so queued requests still show up.
	if s.WaitingReqs == 0 && strings.Contains(body, "num_requests_waiting_by_reason") {
		re := regexp.MustCompile(`(?:vllm:)?num_requests_waiting_by_reason\{[^}]*\}\s+([\d.eE+-]+)`)
		for _, m := range re.FindAllStringSubmatch(body, -1) {
			if v, err := strconv.ParseFloat(m[1], 64); err == nil {
				s.WaitingReqs += v
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
