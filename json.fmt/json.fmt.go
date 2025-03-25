/*
history:
2013-09-06 v1

GoFmt
GoBuildNull
go build -o $HOME/bin/
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	var err error
	var data []byte
	data, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stdin: %v", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, data, "", "	")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json indent: %v", err)
		os.Exit(1)
	}
	err = buf.WriteByte('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "buf write: %v", err)
		os.Exit(1)
	}

	_, err = os.Stdout.Write(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "write to stdout: %v", err)
		os.Exit(1)
	}
}
