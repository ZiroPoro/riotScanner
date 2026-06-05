package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edu/riotscanner/internal/config"
	"github.com/edu/riotscanner/internal/orchestrator"
)

func main() {
	configPath := flag.String("config", "configs/default.yaml", "path to YAML config")
	startEmu := flag.Bool("start-emulator", false, "launch AVD before connecting")
	command := flag.String("cmd", "run", "command: run|clone-create|clone-delete|inject-qr|connect")
	flag.Parse()

	logger := log.New(os.Stdout, "[riotscanner] ", log.LstdFlags)

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

	orch := orchestrator.New(cfg, logger)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	switch *command {
	case "run":
		if err := prepareConnection(ctx, orch, *startEmu); err != nil {
			logger.Fatalf("connect: %v", err)
		}
		defer orch.Stop()
		if err := orch.RunAutomationLoop(ctx); err != nil && err != context.Canceled {
			logger.Fatalf("automation: %v", err)
		}

	case "clone-create":
		if err := prepareConnection(ctx, orch, *startEmu); err != nil {
			logger.Fatalf("connect: %v", err)
		}
		defer orch.Stop()
		if err := orch.CreateClone(); err != nil {
			logger.Fatalf("clone create: %v", err)
		}
		logger.Println("clone created successfully")

	case "clone-delete":
		if err := prepareConnection(ctx, orch, *startEmu); err != nil {
			logger.Fatalf("connect: %v", err)
		}
		defer orch.Stop()
		if err := orch.DeleteClone(); err != nil {
			logger.Fatalf("clone delete: %v", err)
		}
		logger.Println("clone deleted successfully")

	case "inject-qr":
		if err := prepareConnection(ctx, orch, *startEmu); err != nil {
			logger.Fatalf("connect: %v", err)
		}
		defer orch.Stop()
		if err := orch.InjectQR(); err != nil {
			logger.Fatalf("inject qr: %v", err)
		}
		logger.Println("QR image injected into virtual camera")

	case "connect":
		if err := prepareConnection(ctx, orch, *startEmu); err != nil {
			logger.Fatalf("connect: %v", err)
		}
		defer orch.Stop()
		logger.Println("connected; press Ctrl+C to exit")
		<-ctx.Done()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", *command)
		os.Exit(2)
	}
}

func prepareConnection(ctx context.Context, orch *orchestrator.Orchestrator, startEmu bool) error {
	if startEmu {
		return orch.StartEmulator(ctx)
	}
	return orch.ConnectADB()
}
