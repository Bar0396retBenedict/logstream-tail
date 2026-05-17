// Package main is the entry point for the logstream-tail CLI.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/logstream-tail/internal/cli"
	"github.com/user/logstream-tail/internal/formatter"
	"github.com/user/logstream-tail/internal/output"
	"github.com/user/logstream-tail/internal/source"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := cli.ParseFlags(os.Args[1:])
	if err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sources, err := buildSources(cfg)
	if err != nil {
		return fmt.Errorf("building sources: %w", err)
	}

	fmt, err := formatter.New(cfg.Style, cfg.ShowSource)
	if err != nil {
		return fmt.Errorf("creating formatter: %w", err)
	}

	ch := source.NewFanIn(ctx, sources...)
	w := output.New(os.Stdout, fmt)
	return w.Run(ctx, ch)
}

func buildSources(cfg *cli.Config) ([]<-chan logevent.Event, error) {
	var srcs []<-chan logevent.Event

	for _, cw := range cfg.CloudWatch {
		s, err := source.NewCloudWatchSource(cfg.Context, cw)
		if err != nil {
			return nil, fmt.Errorf("cloudwatch source %q: %w", cw.LogGroup, err)
		}
		srcs = append(srcs, s)
	}

	for _, gcp := range cfg.GCP {
		s, err := source.NewGCPSource(cfg.Context, gcp)
		if err != nil {
			return nil, fmt.Errorf("gcp source %q: %w", gcp.LogName, err)
		}
		srcs = append(srcs, s)
	}

	if len(srcs) == 0 {
		return nil, fmt.Errorf("no log sources configured; specify --cw-log-group or --gcp-log-name")
	}

	return srcs, nil
}
