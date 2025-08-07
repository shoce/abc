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
	S1, S2 string
	R1     *regexp.Regexp

	ScannerBuffer []byte
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr,
			"usage: sre S1 [S2]"+NL+
				SPAC+"S1 and S2 are literal strings"+NL+
				"usage: srer R1 [S2]"+NL+
				SPAC+"R1 is a regexp, S2 is a string with $n for submatches"+NL,
		)
		os.Exit(1)
	}
	S1 = os.Args[1]
	if len(os.Args) == 3 {
		S2 = os.Args[2]
	}

	if path.Base(os.Args[0]) == "srer" {
		var err error
		if R1, err = regexp.Compile(S1); err != nil {
			fmt.Fprintf(os.Stderr, "provided regular expression compile error:"+NL+"%v"+NL, err)
			os.Exit(1)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	ScannerBuffer = make([]byte, ScannerBufferSize)
	scanner.Buffer(ScannerBuffer, ScannerBufferSize)

	var line1, line2 string

	if R1 == nil {
		for scanner.Scan() {
			line1 = scanner.Text()
			// https://pkg.go.dev/strings#ReplaceAll
			line2 = strings.ReplaceAll(line1, S1, S2)
			fmt.Println(line2)
		}
	} else {
		for scanner.Scan() {
			line1 = scanner.Text()
			// https://pkg.go.dev/regexp#Regexp.ReplaceAllString
			line2 = R1.ReplaceAllString(line1, S2)
			fmt.Println(line2)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v"+NL, err)
		os.Exit(1)
	}
}
