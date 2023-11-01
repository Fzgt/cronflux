// Command backoff prints the retry delays an exponential policy produces,
// demonstrating the standalone backoff package.
package main

import (
	"fmt"
	"time"

	"github.com/Fzgt/cronflux/backoff"
)

func main() {
	e := backoff.Exponential{
		Base:   200 * time.Millisecond,
		Max:    10 * time.Second,
		Factor: 2,
		Jitter: 0.1,
	}

	fmt.Println("attempt  delay")
	for i := 0; i < 8; i++ {
		fmt.Printf("%5d    %s\n", i, e.Delay(i).Round(time.Millisecond))
	}
}
