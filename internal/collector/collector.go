package collector

import (
	"context"
	"github.com/changye/llmtop/internal/metrics"
)

type GPUCollector interface {
	Name() string
	Fetch(ctx context.Context) (metrics.GPU, error)
}

type NvidiaSMI struct{ ID int }

func NewNvidiaSMI(gpuID int) *NvidiaSMI { return &NvidiaSMI{ID: gpuID} }

func (n *NvidiaSMI) Name() string { return "NVIDIA" }

var _ GPUCollector = (*NvidiaSMI)(nil)
