// GoFixDiff GoFixFix GoGet GoFmt GoBuildNull GoBuild

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	NL = "\n"
)

func main() {
	var err error
	var bb []byte
	bb, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR ReadAll %+v"+NL, err)
		os.Exit(1)
	}
	s := string(bb)
	for strings.Contains(s, `"`) {
		s = strings.Replace(s, `"`, `«`, 1)
		s = strings.Replace(s, `"`, `»`, 1)
	}
	fmt.Print(s)
}
