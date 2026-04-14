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
	TAB = "\t"
	NL  = "\n"
)

var (
	TRASH = "/trash/"
)

func init() {
	if v := os.Getenv("TRASH"); v != "" {
		TRASH = v
	}
}

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
		perr("ERROR invalid command name [%s]", os.Args[0])
		os.Exit(1)
	}
}

func rem() {
	wd, err := os.Getwd()
	if err != nil {
		perr("ERROR Getwd %v", err)
		os.Exit(1)
	}
	for _, a := range os.Args[1:] {
		apath := a
		if !path.IsAbs(apath) {
			apath = path.Join(wd, apath)
		}
		trashapathdir := path.Join(TRASH, path.Dir(apath))
		err = os.MkdirAll(trashapathdir, 0700)
		if err != nil {
			perr("ERROR %v", err)
		}
		trashapath := path.Join(TRASH, apath)
		pout(apath)
		err = os.Rename(apath, trashapath)
		if err != nil {
			perr("ERROR %v", err)
		} else {
			pout(TAB + trashapath)
		}
	}
	os.Exit(0)
}

func remls() {
	pout("lsr %s/", TRASH)
	os.Exit(0)
}

func remrem() {
	pout("rm -r -v %s/*", TRASH)
	os.Exit(0)
}

func perr(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func pout(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stdout, msg+NL)
	} else {
		fmt.Fprintf(os.Stdout, msg+NL, args...)
	}
}
