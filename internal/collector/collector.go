package collector

import (
	"context"
	"github.com/y9c/llmtop/internal/metrics"
)

type GPUCollector interface {
	Name() string
	Fetch(ctx context.Context) ([]metrics.GPU, error)
}

type NVMLCollector struct{}

func NewNVMLCollector() *NVMLCollector { return &NVMLCollector{} }

func (n *NVMLCollector) Name() string { return "NVIDIA" }

var _ GPUCollector = (*NVMLCollector)(nil)
