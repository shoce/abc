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
	
	F = fmt.Sprintf
	pout = fmt.Print
)

func init() {
	var err error
	if v := os.Getenv("TRASH"); v != "" {
		TRASH = v
		TRASH, err = filepath.Abs(TRASH)
		if err != nil {
			perr(F("ERROR filepath.Abs [%s] %v", TRASH, err))
			os.Exit(1)
		}
		TRASH += string(filepath.Separator)
	} else if v := os.Getenv("HOME"); v != "" {
		TRASH = filepath.Join(v, TRASH) + string(filepath.Separator)
	}
	perr(F("DEBUG TRASH [%s]", TRASH))
}

func main() {
	cmdname := filepath.Base(os.Args[0])
	switch cmdname {
	case "rem":
		rem()
	case "remdu":
		remdu()
	case "remls":
		remls()
	case "remrem":
		remrem()
	default:
		perr(F("ERROR invalid command name [%s]", cmdname))
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
		perr("USAGE"+NL+
			TAB+"rem path..."+NL+
			TAB+"remdu"+NL+
			TAB+"remls"+NL+
			TAB+"remrem"+NL,
			)
		os.Exit(1)
	}
	for _, a := range args {
		apath := a
		if !filepath.IsAbs(apath) {
			apath, err = filepath.Abs(apath)
			if err != nil {
				perr(F("ERROR filepath.Abs [%s] %v", apath, err))
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
			perr(F(TAB+"ERROR %v", err))
		}
		trashapath := filepath.Join(TRASH, apath)
		err = os.Rename(apath, trashapath)
		if err != nil {
			perr(F(TAB+"ERROR %v", err))
		} else {
			perr(TAB + trashapath)
		}
	}
	os.Exit(0)
}

func remdu() {
	pout("du -s -m" + SP + TRASH + NL)
	os.Exit(0)
}

func remls() {
	pout("lsr" + SP + TRASH + NL)
	os.Exit(0)
}

func remrem() {
	pout("rm -r -f -v" + SP + TRASH + NL)
	os.Exit(0)
}

func perr(msg string) {
	fmt.Fprint(os.Stderr, msg+NL)
}

