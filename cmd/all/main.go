package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeilh/agentic_go_signals/internal/config"
	"github.com/adeilh/agentic_go_signals/internal/svc"
	"github.com/adeilh/agentic_go_signals/internal/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	if err := svc.Init(cfg); err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return worker.Start(ctx, svc.DB) })
	g.Go(func() error { return svc.App.Listen(":3333") })
	if err := g.Wait(); err != nil {
		panic(err)
	}
}
