/*
GoGet
GoFmt
GoBuildNull
GoBuild
*/

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/rusenask/docker-registry-client/registry"
)

const (
	TAB = "\t"
	NL = "\n"
)

var (
	RegistryUser   string
	RegistryPass   string
	RegistryUrl        string
	RegistryHost       string
	RegistryRepo string
	
	F = fmt.Sprintf
	pout = fmt.Print
)

func init() {
	RegistryUser = os.Getenv("RegistryUser")
	RegistryPass = os.Getenv("RegistryPass")
	/*
		if RegistryUsername == "" {
			perr("WARNING RegistryUser env var empty")
		}
		if RegistryPassword == "" {
			perr("WARNING RegistryPass env var empty")
		}
	*/
}

type Versions []string

func (vv Versions) Len() int {
	return len(vv)
}

func (vv Versions) Less(i, j int) bool {
	v1, v2 := vv[i], vv[j]
	v1s := strings.Split(v1, ".")
	v2s := strings.Split(v2, ".")
	if len(v1s) < len(v2s) {
		return true
	} else if len(v1s) > len(v2s) {
		return false
	}
	for e := 0; e < len(v1s); e++ {
		d1, _ := strconv.Atoi(v1s[e])
		d2, _ := strconv.Atoi(v2s[e])
		if d1 < d2 {
			return true
		} else if d1 > d2 {
			return false
		}
	}
	return false
}

func (vv Versions) Swap(i, j int) {
	vv[i], vv[j] = vv[j], vv[i]
}

func main() {
	all := flag.Bool("all", false, "to print all tags, otherwise only the last tag is printed")
	flag.Parse()

	var args []string
	for _, a := range os.Args[1:] {
		if a != "" {
			args = append(args, a)
		}
	}

	if len(args) < 1 {
		perr(
			"USAGE drlatest docker.registry.repository.url ..."+
			NL+
			"ENV"+
			NL+
			TAB+"RegistryUser"+
			NL+
			TAB+"RegistryPass"+
			NL,
		)
		os.Exit(1)
	}

	for _, a := range args {

		if u, err := url.Parse(a); err != nil {
			perr(F("ERROR [%s] url parse %v", a, err))
			os.Exit(1)
		} else {
			if u.Scheme == "oci" {
				u.Scheme = "https"
			}
			RegistryUrl = F("%s://%s", u.Scheme, u.Host)
			RegistryHost = u.Host
			RegistryRepo = u.Path
		}
		//perr("DEBUG registry [%s] repo [%s]", RegistryUrl, RegistryRepo)

		r := registry.NewInsecure(RegistryUrl, RegistryUser, RegistryPass)
		r.Logf = registry.Quiet

		tags, err := r.Tags(RegistryRepo)
		if err != nil {
			perr(F("ERROR list tags %v", err))
			os.Exit(1)
		}

		sort.Sort(Versions(tags))

		if *all {
			for _, tag := range tags {
					pout(F("%s%s:%s", RegistryHost, RegistryRepo, tag)+NL)
			}
		} else if len(tags) > 0 {
				pout(F("%s%s:%s", RegistryHost, RegistryRepo, tags[len(tags)-1])+NL)
		}

	}
}

func perr(msg string) (int, error) {
	return fmt.Fprint(os.Stderr, msg+NL)
}

