/*
INSTALL
ln -s g gv
ln -s g gr
ln -s g gvr
*/
/*
HISTORY
026/0519 func argss()
*/

// GoGet GoFmt GoBuildNull GoBuild

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	NL   = "\n"
	SPAC = "    "

	ScannerBufferSize = 1 << 20 //ae:>>
)

var (
	PNAME string
	SS    []string
	RR    []*regexp.Regexp

	RegexpMatch = false
	InvertMatch = false

	F = fmt.Sprintf
	pout = fmt.Print
)

func argss() (args []string) {
args = os.Args[1:]

fd3 := os.NewFile(3, "fd3")
if fd3 == nil { perr("ERROR NewFile <3>"); return; }
defer fd3.Close()
data, _ := io.ReadAll(fd3)
if len(data)==0 { return; }
if len(data)==1 && data[0]=='\n' { return; }

// https://pkg.go.dev/strings#Split
args = strings.Split(string(data), NL)
// ae:<
if len(args)>0 && args[len(args)-1]=="" { 
args = args[:len(args)-1] 
}
return

}

func main() {
	PNAME = filepath.Base(os.Args[0])
	SS = argss()
	perr(F("DEBUG SS ([%s])", strings.Join(SS, "][")))

	if len(SS) < 1 {
		switch PNAME {
		case "g":
			perr("USAGE g S" + NL +
				SPAC + "S is a literal string")
		case "gr":
			perr("USAGE gr R" + NL +
				SPAC + "R is a regexp")
		case "gv":
			perr("USAGE gv S" + NL +
				SPAC + "S is a literal string")
		case "gvr":
			perr("USAGE gvr R" + NL +
				SPAC + "R is a regexp")
		}
		os.Exit(1)
	}

	if PNAME == "gv" || PNAME == "gvr" {
		InvertMatch = true
	}

	if PNAME == "gr" || PNAME == "gvr" {
		RegexpMatch = true
		for _, S := range SS {
			R, err := regexp.Compile(S)
			if err != nil {
				perr(F("ERROR regular expression [%s] compile %v", S, err))
				os.Exit(1)
			}
			RR = append(RR, R)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(nil, ScannerBufferSize)

	var line string

	if RegexpMatch {
		// https://pkg.go.dev/regexp#Regexp.MatchString
		for scanner.Scan() {
			line = scanner.Text()
			if InvertMatch {
				match := false
				for _, R := range RR {
					if R.MatchString(line) {
						match = true
					}
				}
				if !match {
					pout(line+NL)
				}
			} else {
				for _, R := range RR {
					if R.MatchString(line) {
						pout(line+NL)
						break
					}
				}
			}
		}
	} else {
		// https://pkg.go.dev/strings#Contains
		for scanner.Scan() {
			line = scanner.Text()
			if InvertMatch {
				match := false
				for _, S := range SS {
					if strings.Contains(line, S) {
						match = true
					}
				}
				if !match {
					pout(line+NL)
				}
			} else {
				for _, S := range SS {
					if strings.Contains(line, S) {
						pout(line+NL)
						break
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		perr(F("ERROR reading input %v", err))
		os.Exit(1)
	}
}

func perr(msgtext string) { fmt.Fprint(os.Stderr, msgtext+NL) }

