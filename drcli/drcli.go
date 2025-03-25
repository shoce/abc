/*
history:
021/0827 v1

go get -u -a -v
go get -u -v github.com/shoce/docker-registry-client/registry@0.1.10
go get -u -v github.com/rusenask/docker-registry-client/registry
GoFmt GoBuild GoRun
GOOS=linux GOARCH=amd64 go build -o $home/build/ .
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rusenask/docker-registry-client/registry"
)

const NL = "\n"

func log(msg string, args ...interface{}) {
	const NL = "\n"
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func main() {
	var url, username, password, rep, manifest string

	flag.Parse()
	if flag.NArg() < 4 {
		log("usage: drcli registry.url username password repository [manifest]")
		os.Exit(1)
	}
	url, username, password, rep = flag.Args()[0], flag.Args()[1], flag.Args()[2], flag.Args()[3]
	if flag.NArg() == 5 {
		manifest = flag.Args()[4]
	}
	r := registry.NewInsecure(url, username, password)
	r.Logf = registry.Quiet
	//log("registry: %v", r)

	tags, err := r.Tags(rep)
	if err != nil {
		log("error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s"+NL, rep, strings.Join(tags, " "))

	if manifest != "" {
		m, err := r.ManifestV2(rep, manifest)
		if err != nil {
			log("error: %v", err)
			os.Exit(1)
		}
		fmt.Printf("manifest: %+v"+NL, m.Config)
	}

	if manifest != "" {
		md, err := r.ManifestDigest(rep, manifest)
		if err != nil {
			log("error: %v", err)
			os.Exit(1)
		}
		fmt.Printf("digest: %+v"+NL, md)
	}
}
