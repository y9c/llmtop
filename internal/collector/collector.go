package collector

import (
	"context"
	"sync"

	"github.com/y9c/llmtop/internal/metrics"
)

type GPUCollector interface {
	Name() string
	Fetch(ctx context.Context) ([]metrics.GPU, error)
}

// NVMLCollector collects GPU metrics via nvml.
// nvml is initialized once (lazily) on the first Fetch and cached.
type NVMLCollector struct {
	mu          sync.Mutex
	initialized bool
}

func NewNVMLCollector() *NVMLCollector { return &NVMLCollector{} }

func (n *NVMLCollector) Name() string { return "NVIDIA" }

var _ GPUCollector = (*NVMLCollector)(nil)
