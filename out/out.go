/*
history:
2020/1227 v1

GoGet GoFmt GoBuild

out ha ha
out 'he he'
err haa >/dev/null
*/

package main

import (
	"fmt"
	"os"
	"path"
)

func main() {
	var err error
	var stream *os.File
	name := path.Base(os.Args[0])
	switch name {
	case "out":
		stream = os.Stdout
	case "err":
		stream = os.Stderr
	default:
		os.Exit(1)
	}
	for _, s := range os.Args[1:] {
		_, err = fmt.Fprintln(stream, s)
		if err != nil {
			os.Exit(1)
		}
	}
}
