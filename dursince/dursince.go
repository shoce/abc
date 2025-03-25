/*
history:
024/0624 v1

go get -a -u -v
go mod tidy

GoFmt
GoBuildNull
GoBuild
GoRun 2022-01-31T16:47:55Z

*/

package main

import (
	"fmt"
	"os"
	"time"
)

const (
	NL = "\n"
)

func main() {

	var err error

	var TimeSince time.Time
	var DurationSince time.Duration

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: dursince iso-8601-timestamp"+NL)
		os.Exit(1)
	}

	TimeSince, err = time.Parse("2006-01-02T15:04:05Z", os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "time.ParseDuration %s: %v"+NL, os.Args[1], err)
		os.Exit(1)
	}

	DurationSince = time.Now().Sub(TimeSince).Round(time.Second)
	fmt.Printf("%v"+NL, DurationSince)

}
