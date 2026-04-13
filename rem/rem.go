/*
history:
2026/0414 v1
*/

// GoGet GoFmt GoBuild

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
	name := path.Base(os.Args[0])
	switch name {
	case "rem":
		rem()
	case "remls":
		remls()
	case "remrem":
		remrem()
	default:
		fmt.Fprintf(os.Stderr, "ERROR invalid command name [%s]"+NL, os.Args[0])
		os.Exit(1)
	}
}

func rem() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR Getwd %v", err)
		os.Exit(1)
	}
	for _, a := range os.Args[1:] {
		apath := a
		if !path.IsAbs(apath) {
			apath = path.Join(wd, apath)
		}
		fmt.Printf(`mkdir -p "/trash/%s"`+NL, path.Dir(apath))
		fmt.Printf(`mv -v "%s" "/trash/%s"`+NL, apath, apath)
	}
	os.Exit(0)
}

func remls() {
	fmt.Printf("lsr /trash/" + NL)
	os.Exit(0)
}

func remrem() {
	fmt.Printf("rm -r -v /trash/*" + NL)
	os.Exit(0)
}
