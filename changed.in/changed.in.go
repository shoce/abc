/*
history:
2020/03/20 v1

GoFmt GoBuildNull GoBuild GoRelease
GoRun 10m .
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	dur time.Duration
)

func changedIn(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if time.Since(info.ModTime()) < dur {
		fmt.Println(path)
	}

	return nil
}

func main() {
	var err error
	var paths []string

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: changed.in N[s|m|h] [path]")
		os.Exit(1)
	}

	dur, err = time.ParseDuration(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse first argument as time duration: %s\n", err)
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
			fmt.Fprintf(os.Stderr, "filepath.Abs: %s\n", err)
			os.Exit(1)
		}

		path = filepath.Clean(path)

		err = filepath.Walk(path, changedIn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "filepath.Walk: %s\n", err)
			os.Exit(1)
		}
	}

	return
}
