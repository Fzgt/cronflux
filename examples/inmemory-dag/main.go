// Command inmemory-dag runs a three-step DAG (extract -> transform -> load)
// entirely in memory and prints the order in which the jobs executed.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/scheduler"
	"github.com/Fzgt/cronflux/internal/store/memory"
)

func main() {
	st := memory.New()

	var mu sync.Mutex
	var order []string
	exec := scheduler.ExecutorFunc(func(_ context.Context, j job.Job, _ job.Run) error {
		mu.Lock()
		order = append(order, j.ID)
		mu.Unlock()
		return nil
	})

	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := scheduler.New(scheduler.Options{
		Store:        st,
		Executor:     exec,
		Workers:      1,
		TickInterval: 25 * time.Millisecond,
		Logger:       quiet,
	})

	jobs := []job.Job{
		{ID: "extract", Name: "Extract", Enabled: true},
		{ID: "transform", Name: "Transform", Enabled: true, DependsOn: []string{"extract"}},
		{ID: "load", Name: "Load", Enabled: true, DependsOn: []string{"transform"}},
	}
	for _, j := range jobs {
		if err := st.PutJob(context.Background(), j); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if _, err := s.Trigger(context.Background(), "extract"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = s.Run(ctx)

	mu.Lock()
	fmt.Println("execution order:", order)
	mu.Unlock()
}
