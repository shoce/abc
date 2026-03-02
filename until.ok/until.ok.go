// GoFmt GoBuildNull GoBuild

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	NL = "\n"
)

var (
	VERSION string

	VERBOSE bool
)

func main() {

	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Print(VERSION + NL)
		os.Exit(0)
	}

	if os.Getenv("VERBOSE") != "" {
		VERBOSE = true
	}

	var err error

	var Duration time.Duration
	var StopAfter time.Duration

	var cmd string
	var args []string
	var Command *exec.Cmd

	if len(os.Args) < 4 || (os.Args[2] != "--" && os.Args[3] != "--") {
		perr("usage: until.ok duration [stopafter] -- command [args]")
		os.Exit(1)
	}

	Duration, err = time.ParseDuration(os.Args[1])
	if err != nil {
		perr("ERROR time.ParseDuration [%s] %v", os.Args[1], err)
		os.Exit(1)
	}

	if os.Args[2] != "--" {
		StopAfter, err = time.ParseDuration(os.Args[2])
		if err != nil {
			perr("ERROR time.ParseDuration [%s] %v", os.Args[2], err)
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
		perr("ERROR there must be `--` before the command")
		os.Exit(1)
	}

	StartTime := time.Now()

	for {
		Command = exec.Command(cmd, args...)
		Command.Stdin, Command.Stdout, Command.Stderr = os.Stdin, os.Stdout, os.Stderr
		perr("VERBOSE %s :", Command)

		err = Command.Run()
		os.Stdout.Sync()
		os.Stderr.Sync()
		if err != nil {
			perr("ERROR %v", err)
		}
		if err == nil {
			os.Exit(0)
		}

		perr("VERBOSE sleeping %v", Duration)
		time.Sleep(Duration)

		perr("VERBOSE passed %v", time.Now().Sub(StartTime).Round(time.Second))
		if StopAfter > 0 && time.Now().Sub(StartTime) > StopAfter {
			perr("VERBOSE stopping after %v passed", StopAfter)
			break
		}
	}

}

func perr(msg string, args ...interface{}) {
	if strings.HasPrefix(msg, "VERBOSE ") && !VERBOSE {
		return
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}
