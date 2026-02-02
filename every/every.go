/*
history:
022/1204 v1
*/

/*
GoFmt GoBuildNull GoBuild
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	NL = "\n"
)

var (
	VERBOSE bool
)

func main() {

	var err error

	var Duration time.Duration
	var StopAfter time.Duration

	var cmd string
	var args []string
	var Command *exec.Cmd

	if len(os.Args) < 4 || (os.Args[2] != "--" && os.Args[3] != "--") {
		fmt.Fprintf(os.Stderr, "usage: every duration [stopafter] -- command [args]"+NL)
		os.Exit(1)
	}

	Duration, err = time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR time.ParseDuration [%s] %v"+NL, os.Args[1], err)
		os.Exit(1)
	}

	if os.Args[2] != "--" {
		StopAfter, err = time.ParseDuration(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR time.ParseDuration [%s] %v"+NL, os.Args[2], err)
			os.Exit(1)
		}
	}

	if os.Args[2] == "--" {
		cmd = os.Args[3]
		args = os.Args[4:]
	} else if os.Args[3] == "--" {
		cmd = os.Args[4]
		args = os.Args[5:]
	} else {
		fmt.Fprintf(os.Stderr, "ERROR there must be '--' before the command"+NL)
		os.Exit(1)
	}

	StartTime := time.Now()

	for {
		Command = exec.Command(cmd, args...)
		Command.Stdin, Command.Stdout, Command.Stderr = os.Stdin, os.Stdout, os.Stderr
		if VERBOSE {
			fmt.Fprintf(os.Stderr, NL+"%s:"+NL, Command)
		}

		err = Command.Run()
		os.Stdout.Sync()
		os.Stderr.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, NL+"ERROR %v"+NL, err)
		}

		if VERBOSE {
			fmt.Fprintf(os.Stderr, NL+"VERBOSE sleeping %v"+NL, Duration)
		}
		time.Sleep(Duration)

		if VERBOSE {
			fmt.Fprintf(os.Stderr, "VERBOSE passed %v"+NL, time.Now().Sub(StartTime).Round(time.Second))
		}
		if StopAfter > 0 && time.Now().Sub(StartTime) > StopAfter {
			if VERBOSE {
				fmt.Fprintf(os.Stderr, NL+"VERBOSE stopping after %v passed"+NL, StopAfter)
			}
			break
		}
	}

}
