// Command cronflux runs the distributed job and cron scheduler.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Fzgt/cronflux/internal/api"
	"github.com/Fzgt/cronflux/internal/buildinfo"
	"github.com/Fzgt/cronflux/internal/config"
	"github.com/Fzgt/cronflux/internal/scheduler"
	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/memory"
	"github.com/Fzgt/cronflux/internal/store/postgres"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-version", "version":
			fmt.Println(buildinfo.Current())
			return
		}
	}
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "cronflux:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load(args)
	if err != nil {
		return err
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	logger.Info("starting cronflux", "version", buildinfo.Version, "backend", string(cfg.Backend))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := openStore(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = st.Close() }()

	if cfg.JobsFile != "" {
		if err := loadJobs(ctx, st, cfg.JobsFile, logger); err != nil {
			return err
		}
	}

	sched := scheduler.New(scheduler.Options{
		Store:        st,
		Executor:     scheduler.ShellExecutor{},
		TickInterval: cfg.TickInterval,
		Workers:      cfg.Workers,
		Lease:        cfg.LeaseDuration,
		Logger:       logger,
	})

	srv := api.NewServer(api.Config{
		Addr:      cfg.ListenAddr,
		Store:     st,
		Scheduler: sched,
		Gatherer:  sched.Metrics().Registry(),
		Logger:    logger,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := sched.Run(ctx); err != nil {
			logger.Error("scheduler exited", "err", err)
		}
	}()

	serveErr := srv.ListenAndServe(ctx, 10*time.Second)
	stop()
	wg.Wait()
	return serveErr
}

// openStore builds the configured storage backend.
func openStore(ctx context.Context, cfg config.Config) (store.Store, error) {
	switch cfg.Backend {
	case config.BackendMemory:
		return memory.New(), nil
	case config.BackendPostgres:
		return postgres.Open(ctx, cfg.DatabaseURL)
	default:
		return nil, fmt.Errorf("unknown backend %q", cfg.Backend)
	}
}

// loadJobs reads the jobs file and upserts every job into the store.
func loadJobs(ctx context.Context, st store.Store, path string, logger *slog.Logger) error {
	jobs, err := config.LoadJobs(path)
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if err := st.PutJob(ctx, j); err != nil {
			return err
		}
	}
	logger.Info("loaded jobs", "count", len(jobs), "file", path)
	return nil
}

// newLogger builds a text slog logger at the requested level.
func newLogger(level string) *slog.Logger {
	var lv slog.Level
	switch level {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lv}))
}
