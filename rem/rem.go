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
	"strings"
)

const (
	SP  = " "
	TAB = "\t"
	NL  = "\n"
)

var (
	TRASH = "/trash/"
)

func init() {
	var err error
	if v := os.Getenv("TRASH"); v != "" {
		TRASH = v
		TRASH, err = filepath.Abs(TRASH)
		if err != nil {
			perr("ERROR filepath.Abs [%s] %v", TRASH, err)
			os.Exit(1)
		}
		TRASH += string(filepath.Separator)
	} else if v := os.Getenv("HOME"); v != "" {
		TRASH = filepath.Join(v, TRASH) + string(filepath.Separator)
	}
	perr("DEBUG TRASH [%s]", TRASH)
}

func main() {
	cmdname := filepath.Base(os.Args[0])
	switch cmdname {
	case "rem":
		rem()
	case "remls":
		remls()
	case "remrem":
		remrem()
	default:
		perr("ERROR invalid command name [%s]", cmdname)
		os.Exit(1)
	}
}

func rem() {
	var err error
	var args []string
	for _, a := range os.Args[1:] {
		if a == "" {
			continue
		}
		args = append(args, a)
	}
	if len(args) == 0 {
		perr("USAGE rem path...")
		os.Exit(1)
	}
	for _, a := range args {
		apath := a
		if !filepath.IsAbs(apath) {
			apath, err = filepath.Abs(apath)
			if err != nil {
				perr("ERROR filepath.Abs [%s] %v", apath, err)
				continue
			}
		}
		perr("rem" + TAB + apath)
		if apath+string(filepath.Separator) == TRASH || strings.HasPrefix(apath, TRASH) {
			perr(TAB + "ERROR TRASH IS TRASH")
			continue
		}
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
	pout("lsr" + SP + TRASH)
	os.Exit(0)
}

func remrem() {
	pout("rm -r -v" + SP + filepath.Join(TRASH, "*"))
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
