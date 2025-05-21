// history:
// 2015-08-06 v1

// usage:
// echo '{"a": {"b": {"c": 1}}}' |jsondel a.b.c

// go build -o jsondel jsondel.go
// go fmt jsondel.go

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	var err error
	var o interface{}
	var b []byte
	b, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read stdin: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(b, &o)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot unmarshal json: %v\n", err)
		os.Exit(1)
	}

	for i, _ := range os.Args[1:] {
		var oi interface{}
		oi = o
		var ss []string
		ss = strings.Split(os.Args[i+1], ".")
		for i, si := range ss[:len(ss)-1] {
			m, ok := oi.(map[string]interface{})
			if !ok {
				fmt.Fprintf(os.Stderr, "object `%v` is not a dict\n", strings.Join(ss[:i], "."))
				os.Exit(1)
			}
			oi, ok = m[si]
			if !ok {
				fmt.Fprintf(os.Stderr, "key=`%s` not found", strings.Join(ss[:i+1], "."))
				os.Exit(1)
			}
		}

	}
	fmt.Printf("%v\n", o)
}
