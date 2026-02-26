package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/HT4w5/autoindex/internal/app"
	"github.com/HT4w5/autoindex/internal/config"
	"github.com/HT4w5/autoindex/internal/meta"
	flag "github.com/spf13/pflag"
)

const (
	exitSuccess = iota
	exitErr
	exitBadConfig
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
		os.Exit(exitSuccess)
	}

	if showVersion {
		fmt.Printf("%s\n", meta.VersionLong)
		os.Exit(exitSuccess)
	}

	// Load configuration
	var cfg config.Config
	var err error

	if configPath != "" {
		err = cfg.LoadFromPath(configPath)
		if err != nil {
			fmt.Printf("error loading configuration from %s: %v\n", configPath, err)
			os.Exit(exitBadConfig)
		}
	} else {
		err = cfg.Load()
		if err != nil {
			fmt.Printf("error loading configuration: %v\n", err)
			os.Exit(exitBadConfig)
		}
	}

	// Validate
	if errs, ok := cfg.Validate(); !ok {
		fmt.Printf("configuration test failed\n")
		for _, v := range errs {
			fmt.Println(v.Error())
		}
		os.Exit(exitBadConfig)
	}

	if testConfig {
		fmt.Printf("configuration test ok\n")
		os.Exit(exitSuccess)
	}

	application := app.New(cfg)

	err = application.Start()
	if err != nil {
		os.Exit(exitErr)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	application.Shutdown()
}
