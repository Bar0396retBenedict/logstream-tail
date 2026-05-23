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
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := cli.ParseFlags(args)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sources, err := buildSources(cfg)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no sources configured; pass --cloudwatch-log-group or --gcp-project")
	}

	fanIn, merged := source.NewFanIn(sources...)
	go fanIn.Run(ctx)

	// Buffer the merged stream to smooth micro-bursts.
	bufCfg := source.DefaultBufferConfig()
	buf, buffered := source.NewBuffer(merged, bufCfg)
	go buf.Run(ctx)

	fmt := formatter.New(formatter.StyleFromString(cfg.Style))
	w := output.New(os.Stdout, fmt)
	return w.Run(ctx, buffered)
}

func buildSources(cfg *cli.Config) ([]source.Source, error) {
	var sources []source.Source

	if cfg.CloudWatchLogGroup != "" {
		s, err := source.NewCloudWatchSource(source.CloudWatchConfig{
			LogGroup:     cfg.CloudWatchLogGroup,
			Region:       cfg.CloudWatchRegion,
			PollInterval: cfg.PollInterval,
			Filter:       source.DefaultConfig(),
		})
		if err != nil {
			return nil, fmt.Errorf("cloudwatch: %w", err)
		}
		sources = append(sources, s)
	}

	if cfg.GCPProject != "" {
		s, err := source.NewGCPSource(source.GCPConfig{
			Project:      cfg.GCPProject,
			LogID:        cfg.GCPLogID,
			PollInterval: cfg.PollInterval,
			Filter:       source.DefaultConfig(),
		})
		if err != nil {
			return nil, fmt.Errorf("gcp: %w", err)
		}
		sources = append(sources, s)
	}

	return sources, nil
}
