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
// gpuID selects a single GPU (0-based); -1 monitors all GPUs.
type NVMLCollector struct {
	mu          sync.Mutex
	initialized bool
	gpuID       int
}

func NewNVMLCollector(gpuID int) *NVMLCollector { return &NVMLCollector{gpuID: gpuID} }

func (n *NVMLCollector) Name() string { return "NVIDIA" }

var _ GPUCollector = (*NVMLCollector)(nil)
