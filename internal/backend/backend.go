package backend

import "github.com/y9c/llmtop/internal/metrics"

type Backend interface {
	Name() string
	Detect(body string) bool
	Parse(body string) (metrics.Snapshot, error)
}

var _ Backend = (*VLLM)(nil)

var backends = []Backend{&SGLang{}, &LLamaCPP{}, &Ollama{}, &VLLM{}}

func Detect(body string) Backend {
	for _, b := range backends {
		if b.Detect(body) {
			return b
		}
	}
	return &VLLM{}
}

var backendByName = map[string]Backend{
	"vllm":    &VLLM{},
	"llamacpp": &LLamaCPP{},
	"ollama":  &Ollama{},
	"sglang":  &SGLang{},
}

// ByName returns a backend by name. Returns nil if not found.
func ByName(name string) Backend {
	return backendByName[name]
}
