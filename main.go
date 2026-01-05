package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bayological/foreman/internal/foreman"
)

func main() {
	configPath := flag.String("config", "configs/foreman.yaml", "path to config file")
	flag.Parse()

	cfg, err := foreman.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	f, err := foreman.New(cfg)
	if err != nil {
		log.Fatalf("failed to create foreman: %v", err)
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	log.Println("Starting Foreman...")
	if err := f.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("foreman exited with error: %v", err)
	}
}