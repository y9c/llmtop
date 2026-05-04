package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Host     string
	Port     int
	Backend  string // auto | vllm | sglang | ollama | llmcpp
	Rate     time.Duration
	GPUID    int
	URL     string
	Scheme  string
	Version string
}

// MetricsURL returns the full metrics URL based on the backend.
func (c *Config) MetricsURL() string {
	if c.URL != "" {
		return c.URL
	}
	path := "/metrics"
	if c.Backend == "ollama" {
		path = "/api/ps"
	}
	return fmt.Sprintf("http://%s:%d%s", c.Host, c.Port, path)
}

// Parse reads flags and env vars, returns Config.
// Default: host=localhost, port=8000, backend=auto, rate=1s, gpu=-1 (all)
func Parse(version string) *Config {
	cfg := &Config{
		Host:    "localhost",
		Port:    8000,
		Backend: "auto",
		Rate:    1 * time.Second,
		GPUID:   -1,
		Version: version,
	}

	// Env overrides
	if v := os.Getenv("LLM_TOP_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("LLM_TOP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := os.Getenv("LLM_TOP_BACKEND"); v != "" {
		cfg.Backend = v
	}
	if v := os.Getenv("LLM_TOP_RATE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Rate = d
		}
	}
	if v := os.Getenv("LLM_TOP_GPU"); v != "" {
		if g, err := strconv.Atoi(v); err == nil {
			cfg.GPUID = g
		}
	}

	// Flag overrides (env already set)
	flag.StringVar(&cfg.URL, "url", cfg.URL, "Full metrics URL (overrides host/port)")
	flag.StringVar(&cfg.Host, "host", cfg.Host, "Metrics host")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Metrics port")
	flag.StringVar(&cfg.Backend, "backend", cfg.Backend, "Backend (auto/vllm/sglang/ollama/llamacpp)")
	flag.DurationVar(&cfg.Rate, "rate", cfg.Rate, "Update interval (e.g. 1s, 500ms)")
	flag.IntVar(&cfg.GPUID, "gpu", cfg.GPUID, "GPU ID (0-based)")
	showHelp := flag.Bool("help", false, "Show help")
	ver := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\nOptions:\n", os.Args[0])
		printFlag("--url", cfg.URL, "Full metrics URL (overrides host/port)")
		printFlag("--host", cfg.Host, "Metrics host")
		printFlag("--port", cfg.Port, "Metrics port")
		printFlag("--backend", cfg.Backend, "Backend (auto/vllm/sglang/ollama/llamacpp)")
		printFlag("--gpu", cfg.GPUID, "GPU ID (0-based)")
		printFlag("--rate", cfg.Rate, "Update interval (e.g. 1s, 500ms)")
		printFlag("--help", false, "Show help")
		printFlag("--version", false, "Show version")
	}

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
	if *ver {
		fmt.Println(cfg.Version)
		os.Exit(0)
	}
	return cfg
}

func printFlag[T any](name string, val T, desc string) {
	fmt.Fprintf(os.Stderr, "  \033[36m%-12s\033[0m %s (\033[33mdefault %v\033[0m)\n", name, desc, val)
}
