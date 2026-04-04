/*
history:
2020/1227 v1

GoGet GoFmt GoBuild

pout ha ha
pout 'he he'
perr haa >/dev/null
*/

package main

import (
	"fmt"
	"os"
	"path"
)

const (
	NL = "\n"
)

func main() {
	var err error
	var stream *os.File
	name := path.Base(os.Args[0])
	switch name {
	case "pout":
		stream = os.Stdout
	case "perr":
		stream = os.Stderr
	default:
		fmt.Fprintf(os.Stderr, "ERROR invalid command name [%s]"+NL, os.Args[0])
		os.Exit(1)
	}
	for _, s := range os.Args[1:] {
		_, err = fmt.Fprintln(stream, s)
		if err != nil {
			os.Exit(1)
		}
	}
}
