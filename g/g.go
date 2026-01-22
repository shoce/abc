/*
GoGet GoFmt GoBuildNull GoBuild
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

const (
	NL   = "\n"
	SPAC = "    "

	ScannerBufferSize = 200 << 10
)

var (
	PNAME string
	SS    []string
	RR    []*regexp.Regexp

	RegexpMatch = false
	InvertMatch = false

	ScannerBuffer []byte
)

func main() {
	if len(os.Args) < 2 {
		perr(
			"usage: g S"+NL+
				SPAC+"S is a literal string"+NL+
				"usage: gr R"+NL+
				SPAC+"R is a regexp"+NL,
			"usage: gv S"+NL+
				SPAC+"S is a literal string"+NL,
			"usage: gvr R"+NL+
				SPAC+"R is a regexp"+NL,
		)
		os.Exit(1)
	}
	PNAME = path.Base(os.Args[0])
	SS = os.Args[1:]

	if PNAME == "gv" || PNAME == "gvr" {
		InvertMatch = true
	}

	if PNAME == "gr" || PNAME == "gvr" {
		RegexpMatch = true
		for _, S := range SS {
			R, err := regexp.Compile(S)
			if err != nil {
				perr("ERROR regular expression [%s] compile %v", S, err)
				os.Exit(1)
			}
			RR = append(RR, R)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	ScannerBuffer = make([]byte, ScannerBufferSize)
	scanner.Buffer(ScannerBuffer, ScannerBufferSize)

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
					pout(line)
				}
			} else {
				for _, R := range RR {
					if R.MatchString(line) {
						pout(line)
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
					pout(line)
				}
			} else {
				for _, S := range SS {
					if strings.Contains(line, S) {
						pout(line)
						break
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		perr("ERROR reading input %v", err)
		os.Exit(1)
	}
}

func perr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+NL, args...)
}

func pout(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, msg+NL, args...)
}
