package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/app"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/config"
	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/logging"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(app.Name, flag.ContinueOnError)
	flags.SetOutput(stderr)

	configPath := flags.String("config", "configs/default.json", "Path to JSON configuration file.")
	showVersion := flags.Bool("version", false, "Print application version and exit.")
	validateConfig := flags.Bool("validate-config", false, "Validate configuration file and exit without starting runtime.")

	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintf(stdout, "%s %s\n", app.Name, app.Version)
		return 0
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "configuration error: %v\n", err)
		return 1
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(stderr, "configuration error: %v\n", err)
		return 1
	}
	if *validateConfig {
		fmt.Fprintf(stdout, "configuration is valid: %s\n", *configPath)
		return 0
	}

	logger := logging.NewTextLogger("info", stdout)
	logger.Info("configuration loaded", "config", *configPath)
	runtime := app.New(cfg, logger)
	if err := runtime.Run(context.Background()); err != nil {
		fmt.Fprintf(stderr, "application error: %v\n", err)
		return 1
	}

	return 0
}
