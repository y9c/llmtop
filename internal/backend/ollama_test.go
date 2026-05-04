package backend

import "testing"

var ollamaSample = `{"models":[{"name":"tinyllama:latest","size_vram":1228800000,"details":{"parameter_size":"1.1B"}}]}`

var ollamaEmptySample = `{"models":[]}`

func TestOllamaName(t *testing.T) {
	var b Ollama
	if got := b.Name(); got != "Ollama" {
		t.Fatalf("Name(): want Ollama, got %s", got)
	}
}

func TestOllamaDetect(t *testing.T) {
	var b Ollama
	if !b.Detect(ollamaSample) {
		t.Fatal("Detect(models+size_vram) should be true")
	}
	if b.Detect("other") {
		t.Fatal("Detect(other) should be false")
	}
}

func TestOllamaParse(t *testing.T) {
	var b Ollama
	s, err := b.Parse(ollamaSample)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if s.GPUUsedMB != 1171.875 {
		t.Fatalf("GPUUsedMB: want 1171.875, got %v", s.GPUUsedMB)
	}
	if s.RunningReqs != 1 {
		t.Fatalf("RunningReqs: want 1, got %v", s.RunningReqs)
	}
}

func TestOllamaParseEmpty(t *testing.T) {
	var b Ollama
	s, err := b.Parse(ollamaEmptySample)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if s.GPUUsedMB != 0 {
		t.Fatalf("GPUUsedMB: want 0 for empty, got %v", s.GPUUsedMB)
	}
	if s.RunningReqs != 0 {
		t.Fatalf("RunningReqs: want 0 for empty, got %v", s.RunningReqs)
	}
}
