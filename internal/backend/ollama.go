package backend

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/y9c/llmtop/internal/metrics"
)

type Ollama struct{}

func (Ollama) Name() string { return "Ollama" }

func (Ollama) Detect(body string) bool {
	return strings.Contains(body, `"models"`) && strings.Contains(body, `"size_vram"`)
}

type ollamaModel struct {
	Name      string  `json:"name"`
	SizeVRAM  float64 `json:"size_vram"`
}

type ollamaPSResponse struct {
	Models []ollamaModel `json:"models"`
}

func (Ollama) Parse(body string) (metrics.Snapshot, error) {
	s := metrics.Snapshot{Timestamp: time.Now()}

	var resp ollamaPSResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return s, err
	}

	var totalVRAM float64
	for _, m := range resp.Models {
		totalVRAM += m.SizeVRAM
	}
	s.RunningReqs = float64(len(resp.Models))

	if totalVRAM > 0 {
		s.GPUUsedMB = totalVRAM / 1024 / 1024
		s.GPUMemTotalMB = s.GPUUsedMB * 2
	}

	return s, nil
}
