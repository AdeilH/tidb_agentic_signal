package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeilh/agentic_go_signals/internal/config"
	"github.com/adeilh/agentic_go_signals/internal/db"
	"github.com/adeilh/agentic_go_signals/internal/svc"
	"github.com/adeilh/agentic_go_signals/internal/worker"
	"golang.org/x/sync/errgroup"
)

func main() {
	// Parse command line flags
	migrateOnly := flag.Bool("migrate-only", false, "Run database migrations only and exit")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// If migrate-only flag is set, just run migrations and exit
	if *migrateOnly {
		fmt.Println("Running database migrations only...")
		database, err := db.Open(cfg.DBDSN)
		if err != nil {
			fmt.Printf("Failed to connect to database: %v\n", err)
			os.Exit(1)
		}
		defer database.GetConn().Close()

		if err := db.AutoMigrate(database); err != nil {
			fmt.Printf("Failed to run migrations: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Database migrations completed successfully!")
		return
	}

	// Normal startup
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
