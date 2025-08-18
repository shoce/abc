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
	S     string
	R     *regexp.Regexp

	InvertMatch   bool
	ScannerBuffer []byte
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr,
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
	S = os.Args[1]

	if PNAME == "gv" || PNAME == "gvr" {
		InvertMatch = true
	}

	if PNAME == "gr" || PNAME == "gvr" {
		var err error
		R, err = regexp.Compile(S)
		if err != nil {
			fmt.Fprintf(os.Stderr, "provided regular expression compile error:"+NL+"%v"+NL, err)
			os.Exit(1)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	ScannerBuffer = make([]byte, ScannerBufferSize)
	scanner.Buffer(ScannerBuffer, ScannerBufferSize)

	var line string

	if R == nil {
		for scanner.Scan() {
			line = scanner.Text()
			// https://pkg.go.dev/strings#Contains
			if strings.Contains(line, S) != InvertMatch {
				fmt.Println(line)
			}
		}
	} else {
		for scanner.Scan() {
			line = scanner.Text()
			// https://pkg.go.dev/regexp#Regexp.MatchString
			if R.MatchString(line) != InvertMatch {
				fmt.Println(line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v"+NL, err)
		os.Exit(1)
	}
}
