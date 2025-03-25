/*
history:
20/1212 v1

usage:
get get-file-test

GoFmt GoBuildNull GoRelease GoBuild
*/

package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func log(msg string, args ...interface{}) {
	const Beat = time.Duration(24) * time.Hour / 1000
	tzBiel := time.FixedZone("Biel", 60*60)
	t := time.Now().In(tzBiel)
	ty := t.Sub(time.Date(t.Year(), 1, 1, 0, 0, 0, 0, tzBiel))
	td := t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tzBiel))
	ts := fmt.Sprintf(
		"%d/%d@%d",
		t.Year()%1000,
		int(ty/(time.Duration(24)*time.Hour))+1,
		int(td/Beat),
	)
	fmt.Fprintf(os.Stderr, ts+" "+msg+"\n", args...)
}

func main() {
	var err error
	var path string

	if len(os.Args) > 1 {
		path = os.Args[1]
	} else {
		log("usage: get path")
		os.Exit(1)
	}

	var f *os.File
	f, err = os.Open(path)
	if err != nil {
		os.Exit(0)
	}
	defer f.Close()

	var s os.FileInfo
	s, err = f.Stat()
	if err != nil {
		log("%#v", err)
		os.Exit(1)
	}
	if s.IsDir() {
		os.Exit(0)
	}

	_, err = io.Copy(os.Stdout, f)
	if err != nil {
		log("%#v", err)
		os.Exit(1)
	}
}
