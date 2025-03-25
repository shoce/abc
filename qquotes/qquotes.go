package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	var err error
	var bb []byte
	bb, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ReadAll:", err)
		os.Exit(1)
	}
	s := string(bb)
	for strings.Contains(s, `"`) {
		s = strings.Replace(s, `"`, `«`, 1)
		s = strings.Replace(s, `"`, `»`, 1)
	}
	fmt.Print(s)
}
