/*
history:
2020/3/31 v1

GoGet
GoFmt
GoBuildNull
GoBuild
*/

package main

import (
	"encoding/base32"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
)

func main() {
	bb, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	e := base32.NewEncoding(crockford)
	// TODO io.Copy
	// https://pkg.go.dev/encoding/base32/
	s := e.EncodeToString(bb)
	fmt.Println(s)
}
