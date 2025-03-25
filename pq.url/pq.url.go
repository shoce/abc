package main

import (
	"fmt"
	"github.com/lib/pq"
	"os"
)

func main() {
	var err error
	var dbparams string

	if len(os.Args) < 2 {
		fmt.Println("usage: $0 postgres://...")
		return
	}
	dbparams, err = pq.ParseURL(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(dbparams)
}
