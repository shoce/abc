/*
usage: sre S1 [S2]
S1 can be a regex, if begins with `~`
replace all occurences of S1 with S2 in stdin and print to stdout
if S2 is not provided or empty, all occurences of S1 are removed

GoFmt
GoBuildNull
GoBuild
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	NL   = "\n"
	SPAC = "    "
)

var (
	S1, S2 string
	R1     *regexp.Regexp
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr,
			"usage: sre S1 [S2]"+NL+
				SPAC+"S1 is regexp if begins with `~`"+NL,
		)
		os.Exit(1)
	}
	S1 = os.Args[1]
	if len(os.Args) == 3 {
		S2 = os.Args[2]
	}

	if strings.HasPrefix(S1, "~") {
		s1 := strings.TrimPrefix(S1, "~")
		var err error
		R1, err = regexp.Compile(s1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "provided regular expression compile error:"+NL+"%v"+NL, err)
			os.Exit(1)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	var line1, line2 string
	for scanner.Scan() {
		line1 = scanner.Text()
		if R1 == nil {
			line2 = strings.ReplaceAll(line1, S1, S2)
		} else {
			// https://pkg.go.dev/regexp#Regexp.ReplaceAllLiteralString
			// https://pkg.go.dev/regexp#Regexp.ReplaceAllString

			line2 = R1.ReplaceAllLiteralString(line1, S2)
		}
		fmt.Println(line2)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v"+NL, err)
		os.Exit(1)
	}
}
