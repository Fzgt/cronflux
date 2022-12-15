// Command cronflux runs the distributed job and cron scheduler.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Fzgt/cronflux/internal/buildinfo"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(buildinfo.Current())
		return
	}

	// TODO: wire the store, scheduler and HTTP server together.
	fmt.Fprintln(os.Stderr, "cronflux: server wiring not implemented yet")
	os.Exit(1)
}
