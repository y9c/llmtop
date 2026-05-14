package collector

import (
	"context"
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/y9c/llmtop/internal/metrics"
)

func (n *NVMLCollector) Fetch(ctx context.Context) ([]metrics.GPU, error) {
	if err := n.ensureInit(ctx); err != nil {
		return nil, err
	}

	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return nil, fmt.Errorf("nvml.DeviceGetCount: %v", ret)
	}
	if count == 0 {
		return nil, fmt.Errorf("nvml: no GPUs found")
	}

	gpus := make([]metrics.GPU, 0, count)
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		dev, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			return nil, fmt.Errorf("nvml.DeviceGetHandleByIndex(%d): %v", i, ret)
		}

		var g metrics.GPU

		name, ret := dev.GetName()
		if ret == nvml.SUCCESS {
			g.Name = name
		}

		mem, ret := dev.GetMemoryInfo()
		if ret == nvml.SUCCESS {
			g.TotalMB = float64(mem.Total) / 1024 / 1024
			g.UsedMB = float64(mem.Used) / 1024 / 1024
		}

		util, ret := dev.GetUtilizationRates()
		if ret == nvml.SUCCESS {
			g.UtilPct = float64(util.Gpu)
		}

		temp, ret := dev.GetTemperature(nvml.TEMPERATURE_GPU)
		if ret == nvml.SUCCESS {
			g.TempC = float64(temp)
		}

		power, ret := dev.GetPowerUsage()
		if ret == nvml.SUCCESS {
			g.PowerW = float64(power) / 1000.0
		}

		powerLimit, ret := dev.GetEnforcedPowerLimit()
		if ret == nvml.SUCCESS {
			g.PowerMaxW = float64(powerLimit) / 1000.0
		}

		gpus = append(gpus, g)
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("nvml: no GPUs found")
	}

	return gpus, nil
}

func (n *NVMLCollector) ensureInit(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.initialized {
		return nil
	}
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("nvml.Init: %v", ret)
	}
	n.initialized = true
	return nil
}

func (n *NVMLCollector) Close() {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.initialized {
		nvml.Shutdown()
		n.initialized = false
	}
}
