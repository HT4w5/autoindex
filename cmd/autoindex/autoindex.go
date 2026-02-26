package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HT4w5/autoindex/internal/app"
	"github.com/HT4w5/autoindex/internal/config"
	"github.com/HT4w5/autoindex/internal/meta"
	flag "github.com/spf13/pflag"
)

func main() {
	var configPath string // Path to configuration file
	var testConfig bool   // Test config and exit
	var showVersion bool  // Show version information
	var showHelp bool     // Show help message

	flag.StringVarP(&configPath, "config", "c", "", "path to configuration file")
	flag.BoolVarP(&showVersion, "version", "v", false, "show version information")
	flag.BoolVarP(&showHelp, "help", "h", false, "show help message")
	flag.BoolVarP(&testConfig, "test", "t", false, "test config and exit")
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
	if errs, ok := cfg.Validate(); !ok {
		log.Fatalf("Configuration validation failed: %v", errs)
	}

	if testConfig {
		os.Exit(0)
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
