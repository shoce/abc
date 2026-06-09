/*
HISTORY
021/0827 v1

go get -u -a -v
go get -u -v github.com/shoce/docker-registry-client/registry@0.1.10
go get -u -v github.com/rusenask/docker-registry-client/registry

GoGet
GoFmt 
GoBuild 
GoRun
*/

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rusenask/docker-registry-client/registry"
)

const (
	NL = "\n"
)

var (
	F = fmt.Sprintf
	pout = fmt.Print
)

func main() {
	args := os.Args[1:]
	if len(args) < 4 {
		perr("USAGE drcli registry.url username password repository [manifest]")
		os.Exit(1)
	}
	
	drurl, user, pass, repo := args[0], args[1], args[2], args[3]
	manifest := ""
	if len(args) == 5 {
		manifest = args[4]
	}
	
	dr := registry.NewInsecure(drurl, user, pass)
	dr.Logf = registry.Quiet
	//perr(F("DEBUG registry %v", dr))

	tags, err := dr.Tags(repo)
	if err != nil {
		perr(F("ERROR %v", err))
		os.Exit(1)
	}
	pout(F("%s %s", repo, strings.Join(tags, " "))+NL)

	if manifest != "" {
		m, err := dr.ManifestV2(repo, manifest)
		if err != nil {
			perr(F("ERROR %v", err))
			os.Exit(1)
		}
		pout(F("manifest %+v", m.Config)+NL)
	}

	if manifest != "" {
		md, err := dr.ManifestDigest(repo, manifest)
		if err != nil {
			perr(F("ERROR %v", err))
			os.Exit(1)
		}
		pout(F("digest %+v", md)+NL)
	}
}

func perr(msg string) (int, error) {
	return fmt.Fprint(os.Stderr, msg+NL)
}

