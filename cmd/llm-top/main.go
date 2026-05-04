package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/y9c/llmtop/internal/app"
	"github.com/y9c/llmtop/internal/collector"
	"github.com/y9c/llmtop/internal/config"
	"github.com/y9c/llmtop/internal/fetcher"
	"github.com/y9c/llmtop/internal/ui"
)

var version = "dev"

func main() {
	// Reduce GC frequency — low allocation rate (~3KB/s), 1Hz tick
	debug.SetGCPercent(200)

	cfg := config.Parse(version)
	f := fetcher.New(5*time.Second, 3)
	gpu := collector.NewNVMLCollector()

	model := &ui.Model{}
	u := app.New(cfg, f, gpu, model)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := u.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
