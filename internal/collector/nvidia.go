package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/y9c/llmtop/internal/metrics"
)

func (n *NvidiaSMI) Fetch(ctx context.Context) ([]metrics.GPU, error) {
	args := []string{
		"--query-gpu=name,memory.used,memory.total,utilization.gpu,temperature.gpu,power.draw,power.limit",
		"--format=csv,noheader,nounits",
	}
	cmd := exec.CommandContext(ctx, "nvidia-smi", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("nvidia-smi: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var gpus []metrics.GPU
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ", ")
		if len(parts) != 7 {
			return nil, fmt.Errorf("nvidia-smi: unexpected output: %q", line)
		}
		used, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("nvidia-smi: parse memory.used: %w", err)
		}
		total, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return nil, fmt.Errorf("nvidia-smi: parse memory.total: %w", err)
		}
		util, err := strconv.ParseFloat(parts[3], 64)
		if err != nil {
			return nil, fmt.Errorf("nvidia-smi: parse utilization.gpu: %w", err)
		}
		temp, err := strconv.ParseFloat(parts[4], 64)
		if err != nil {
			return nil, fmt.Errorf("nvidia-smi: parse temperature.gpu: %w", err)
		}
		power, err := strconv.ParseFloat(parts[5], 64)
		if err != nil {
			return nil, fmt.Errorf("nvidia-smi: parse power.draw: %w", err)
		}
		powerMax, err := strconv.ParseFloat(parts[6], 64)
		if err != nil {
			return nil, fmt.Errorf("nvidia-smi: parse power.limit: %w", err)
		}
		gpus = append(gpus, metrics.GPU{
			Name:      parts[0],
			UsedMB:    used,
			TotalMB:   total,
			UtilPct:   util,
			TempC:     temp,
			PowerW:    power,
			PowerMaxW: powerMax,
		})
	}
	if len(gpus) == 0 {
		return nil, fmt.Errorf("nvidia-smi: no GPUs found")
	}
	return gpus, nil
}
