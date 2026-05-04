package collector

import (
	"context"
	"github.com/y9c/llmtop/internal/metrics"
)

type GPUCollector interface {
	Name() string
	Fetch(ctx context.Context) ([]metrics.GPU, error)
}

type NvidiaSMI struct{}

func NewNvidiaSMI() *NvidiaSMI { return &NvidiaSMI{} }

func (n *NvidiaSMI) Name() string { return "NVIDIA" }

var _ GPUCollector = (*NvidiaSMI)(nil)
