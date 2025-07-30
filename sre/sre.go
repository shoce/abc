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
)

var (
	S1, S2 string
	R1, R2 *regexp.Regexp
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr,
			"usage: sre S1 [S2]"+NL+
				SPAC+"S1 and S2 are literal strings"+NL+
				"usage: srer R1 [R2]"+NL+
				SPAC+"R1 and R2 are regexps"+NL,
		)
		os.Exit(1)
	}
	S1 = os.Args[1]
	if len(os.Args) == 3 {
		S2 = os.Args[2]
	}

	if path.Base(os.Args[0]) == "srer" {
		var err error
		R1, err = regexp.Compile(S1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "provided regular expression compile error:"+NL+"%v"+NL, err)
			os.Exit(1)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
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
			// https://pkg.go.dev/regexp#Regexp.ReplaceAllLiteralString
			// https://pkg.go.dev/regexp#Regexp.ReplaceAllString
			line2 = R1.ReplaceAllLiteralString(line1, S2)
			fmt.Println(line2)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v"+NL, err)
		os.Exit(1)
	}
}
