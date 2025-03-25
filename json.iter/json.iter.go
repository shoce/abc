/*
history:
2015-05-29 v1

usage:
echo '[1,2,3]' |json.iter

GoFmt GoBuildNull GoBuild GoRelease
*/

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	var oo []interface{}
	oo, ok := o.([]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "object is not a list\n")
		os.Exit(1)
	}
	var oi interface{}
	var oib []byte
	for _, oi = range oo {
		oib, err = json.Marshal(oi)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot marshal object `%v` to json\n", oi)
			os.Exit(1)
		}
		fmt.Println(string(oib))
	}
}
