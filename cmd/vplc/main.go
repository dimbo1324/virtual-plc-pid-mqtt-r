package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	runRuntime := flags.Bool("run", false, "Run the PLC runtime until Ctrl+C.")

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

	logWriter := buildLogWriter(stdout, cfg)
	logger := logging.NewTextLogger("info", logWriter)
	logger.Info("configuration loaded", "config", *configPath)
	runtime := app.New(cfg, logger)
	if !*runRuntime {
		if err := runtime.Run(context.Background()); err != nil {
			fmt.Fprintf(stderr, "application error: %v\n", err)
			return 1
		}
		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runtime.RunRuntime(ctx); err != nil {
		fmt.Fprintf(stderr, "application error: %v\n", err)
		return 1
	}
	return 0
}

// buildLogWriter returns a writer that tees stdout to cfg.Storage.AppLogPath when storage
// is enabled and app_log_path is set. Falls back to stdout-only on any file error.
func buildLogWriter(stdout io.Writer, cfg config.Config) io.Writer {
	if !cfg.Storage.Enabled || cfg.Storage.AppLogPath == "" {
		return stdout
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Storage.AppLogPath), 0o750); err != nil {
		return stdout
	}
	f, err := os.OpenFile(cfg.Storage.AppLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return stdout
	}
	// f is intentionally not closed here; it lives for the entire process lifetime.
	return io.MultiWriter(stdout, f)
}
