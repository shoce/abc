/*
HISTORY
2026/0414 v1
*/

// GoGet GoFmt GoBuild

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	TAB = "\t"
	NL  = "\n"
)

var (
	TRASH   = "/trash/"
	CmdName string
)

func init() {
	if v := os.Getenv("TRASH"); v != "" {
		TRASH = v
	} else if v := os.Getenv("HOME"); v != "" {
		TRASH = filepath.Join(v, TRASH) + "/"
	}
	perr("DEBUG TRASH [%s]", TRASH)
}

func main() {
	CmdName = filepath.Base(os.Args[0])
	switch CmdName {
	case "rem":
		rem()
	case "remls":
		remls()
	case "remrem":
		remrem()
	default:
		perr("ERROR invalid command name [%s]", CmdName)
		os.Exit(1)
	}
}

func rem() {
	wd, err := os.Getwd()
	if err != nil {
		perr("ERROR Getwd %v", err)
		os.Exit(1)
	}
	var args []string
	for _, a := range os.Args[1:] {
		if a == "" {
			continue
		}
		args = append(args, a)
	}
	if len(args) == 0 {
		perr("USAGE %s path...", CmdName)
		os.Exit(1)
	}
	for _, a := range args {
		apath := a
		if !filepath.IsAbs(apath) {
			apath = filepath.Join(wd, apath)
		}
		perr(apath)
		trashapathdir := filepath.Join(TRASH, filepath.Dir(apath))
		err = os.MkdirAll(trashapathdir, 0700)
		if err != nil {
			perr(TAB+"ERROR %v", err)
		}
		trashapath := filepath.Join(TRASH, apath)
		err = os.Rename(apath, trashapath)
		if err != nil {
			perr(TAB+"ERROR %v", err)
		} else {
			perr(TAB + trashapath)
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
