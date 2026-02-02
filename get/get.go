/*
history:
20/1212 v1
*/

/*
usage:
get get-file-test

GoFmt GoBuildNull GoBuild
*/

package main

import (
	"fmt"
	"io"
	"os"
)

const (
	NL = "\n"
)

func log(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func main() {
	var err error
	var path string

	if len(os.Args) == 2 {
		path = os.Args[1]
	} else {
		log("usage: get path")
		os.Exit(1)
	}

	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		log("ERROR %v", err)
		os.Exit(1)
	}
	defer f.Close()

	var s os.FileInfo
	s, err = f.Stat()
	if err != nil {
		log("ERROR %v", err)
		os.Exit(1)
	}
	if s.IsDir() {
		log("ERROR %s is a directory", path)
		os.Exit(1)
	}

	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		log("ERROR %v", err)
		os.Exit(1)
	}
}
