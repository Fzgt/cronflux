// Command cron-next prints the next few activation times for a cron spec,
// demonstrating the standalone cron package.
//
// Usage:
//
//	go run ./examples/cron-next "0 9 * * 1-5"
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Fzgt/cronflux/cron"
)

func main() {
	spec := "*/15 * * * *"
	if len(os.Args) > 1 {
		spec = os.Args[1]
	}

	sched, err := cron.Parse(spec)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	t := time.Now()
	fmt.Printf("next fire times for %q:\n", spec)
	for i := 0; i < 5; i++ {
		t = sched.Next(t)
		fmt.Printf("  %s\n", t.Format(time.RFC3339))
	}
}
