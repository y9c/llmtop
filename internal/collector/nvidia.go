package collector

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/changye/llmtop/internal/metrics"
)

func (n *NvidiaSMI) Fetch(ctx context.Context) (metrics.GPU, error) {
	args := []string{
		"--query-gpu=name,memory.used,memory.total,utilization.gpu",
		"--format=csv,noheader,nounits",
	}
	if n.ID >= 0 {
		args = append(args, fmt.Sprintf("--id=%d", n.ID))
	}
	cmd := exec.CommandContext(ctx, "nvidia-smi", args...)
	out, err := cmd.Output()
	if err != nil {
		return metrics.GPU{}, fmt.Errorf("nvidia-smi: %w", err)
	}

	// Handle multi-GPU: take the first line if multiple lines returned
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	line := strings.TrimSpace(lines[0])
	parts := strings.Split(line, ", ")
	if len(parts) != 4 {
		return metrics.GPU{}, fmt.Errorf("nvidia-smi: unexpected output: %q", line)
	}

	used, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return metrics.GPU{}, fmt.Errorf("nvidia-smi: parse memory.used: %w", err)
	}
	total, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return metrics.GPU{}, fmt.Errorf("nvidia-smi: parse memory.total: %w", err)
	}
	util, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return metrics.GPU{}, fmt.Errorf("nvidia-smi: parse utilization.gpu: %w", err)
	}

	return metrics.GPU{
		Name:    parts[0],
		UsedMB:  used,
		TotalMB: total,
		UtilPct: util,
	}, nil
}
