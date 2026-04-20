/*
history:
2020/03/20 v1
*/

// GoFmt GoBuildNull GoBuild GoRelease
// GoRun 10m .

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	NL = "\n"
)

var (
	Dur time.Duration
)

func changedIn(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if time.Since(info.ModTime()) < Dur {
		fmt.Print(path + NL)
	}

	return nil
}

func main() {
	var err error
	var paths []string

	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "USAGE changed.in N[s|m|h] [path]"+NL)
		os.Exit(1)
	}

	Dur, err = time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse first argument [%s] as time duration %v"+NL, os.Args[1], err)
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		paths = os.Args[2:]
	} else {
		paths = []string{"."}
	}

	for _, path := range paths {
		path, err = filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR filepath.Abs %v"+NL, err)
			os.Exit(1)
		}

		path = filepath.Clean(path)

		err = filepath.Walk(path, changedIn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR filepath.Walk %v"+NL, err)
			os.Exit(1)
		}
	}

	return
}
