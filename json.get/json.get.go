/*
history:
2015-04-27 v1
2015-05-01 v2
2015-10-06 v3 output json, not "%v"

usage:
echo '{"a": {"b": {"c": 1}}}' |json.get a.b.c

GoFmt GoBuildNull GoBuild GoRelease
*/

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
		for i, si := range ss {
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
		var oib []byte
		oib, err = json.Marshal(oi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshal json: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(oib))
	}
}
