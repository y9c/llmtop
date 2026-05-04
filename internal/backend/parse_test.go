package backend

import (
    "os"
    "testing"
)

func TestParseRealVLLMMetrics(t *testing.T) {
    body, err := os.ReadFile("/tmp/vllm_metrics_final.txt")
    if err != nil {
        t.Skip("no captured metrics file")
    }
    
    b := VLLM{}
    if !b.Detect(string(body)) {
        t.Fatal("should detect vLLM metrics")
    }
    
    snap, err := b.Parse(string(body))
    if err != nil {
        t.Fatal("parse error:", err)
    }
    
    t.Logf("RunningReqs: %.0f", snap.RunningReqs)
    t.Logf("WaitingReqs: %.0f", snap.WaitingReqs)
    t.Logf("GenTokensTotal: %.0f", snap.GenTokensTotal)
    t.Logf("PromptTokensTotal: %.0f", snap.PromptTokensTotal)
    t.Logf("SpecDraftsTotal: %.0f", snap.SpecDraftsTotal)
    t.Logf("SpecDraftToksTotal: %.0f", snap.SpecDraftToksTotal)
    t.Logf("SpecAcceptedTotal: %.0f", snap.SpecAcceptedTotal)
    t.Logf("PrefixCacheHits: %.0f", snap.PrefixCacheHits)
    t.Logf("PrefixCacheQueries: %.0f", snap.PrefixCacheQueries)
    t.Logf("KVCacheUsagePct: %.4f", snap.KVCacheUsagePct)
    t.Logf("TTFTTotalS: %.4f", snap.TTFTTotalS)
    t.Logf("TTFTCount: %.0f", snap.TTFTCount)
    t.Logf("TPOTTotalS: %.4f", snap.TPOTTotalS)
    t.Logf("TPOTCount: %.0f", snap.TPOTCount)
    t.Logf("StartTimeUnix: %.0f", snap.StartTimeUnix)
    t.Logf("SpecAcceptedPos: %v", snap.SpecAcceptedPos)
    
    // Verify key fields are parsed
    if snap.GenTokensTotal == 0 {
        t.Error("GenTokensTotal should not be 0")
    }
    if snap.TTFTCount == 0 {
        t.Error("TTFTCount should not be 0")
    }
}
