package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HT4w5/index/internal/app"
	"github.com/HT4w5/index/internal/config"
	"github.com/HT4w5/index/internal/meta"
)

func main() {
	var configPath string
	var showVersion bool
	var showHelp bool

	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.StringVar(&configPath, "c", "", "Path to configuration file")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if showVersion {
		fmt.Printf("%s\n", meta.VersionShort)
		os.Exit(0)
	}

	// Load configuration
	var cfg config.Config
	var err error

	if configPath != "" {
		err = cfg.LoadFromPath(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
		}
	} else {
		err = cfg.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Validate
	if msgs, ok := cfg.Validate(); !ok {
		log.Fatalf("Configuration validation failed: %v", msgs)
	}

	application := app.New(cfg)

	err = application.Start()
	if err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan bool, 1)
	go func() {
		err := application.Shutdown()
		if err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		done <- true
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Println("Shutdown timed out")
	}
}
