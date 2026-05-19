/*
HISTORY
024/0624 v1
*/

// GoFmt GoBuildNull GoBuild
// GoRun 2022-01-31T16:47:55Z


package main

import (
	"fmt"
	"os"
	"time"
)

const (
	NL = "\n"
)

var (
	TimeSince time.Time
	DurationSince time.Duration

	F = fmt.Sprintf
	pout = fmt.Print
)

func main() {

	var err error

	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, "USAGE dursince iso-8601-timestamp"+NL)
		os.Exit(1)
	}

	TimeSince, err = time.Parse("2006-01-02T15:04:05Z", args[0])
	if err != nil {
		fmt.Fprint(os.Stderr, F("ERROR time.ParseDuration [%s] %v"+NL, args[0], err))
		os.Exit(1)
	}

	DurationSince = time.Now().Sub(TimeSince).Round(time.Second)
	pout(F("%v"+NL, DurationSince))

}
