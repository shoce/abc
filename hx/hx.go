/*
history:
26/0330@thailand v1
*/

// GoGet GoFmt GoBuildNull
// GoBuild GoRun

package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

const (
	N   = ""
	SP  = " "
	TAB = "\t"
	NL  = "\n"
	SEP = ","
)

var (
	VERSION string
	USAGE   string = `man hx:
hx get scheme://host:port/path/subpath head1:v1 head2:v2 arg1=val1 arg2=val2
`
	DEBUG bool

	HxHeaders bool

	// TODO basic auth
	HxUser string
	HxPass string

	// TODO timeout
	HxTimeout time.Duration

	HxInsecure bool
)

func init() {
	if os.Getenv("DEBUG") != "" {
		DEBUG = true
	}

	if os.Getenv("HxHeaders") != "" {
		HxHeaders = true
	}

	if os.Getenv("HxInsecure") != "" {
		HxInsecure = true
	}

}

func main() {
	var err error

	args := os.Args[1:]
	//perr("DEBUG args %#v", args)
	n := 0
	for _, a := range args {
		if a != "" {
			args[n] = a
			n++
		}
	}
	args = args[:n]
	//perr("DEBUG n <%d> args %#v", n, args)

	if len(args) == 1 && args[0] == "version" {
		fmt.Print(VERSION + NL)
		os.Exit(0)
	}

	if len(args) == 1 && slices.Contains([]string{"help", "usage"}, args[0]) {
		pout(USAGE + NL)
		os.Exit(0)
	}

	if len(args) < 1 {
		perr("ERROR not enough arguments, see hx help/usage")
		os.Exit(1)
	}

	// https://pkg.go.dev/net/http
	var hmethod string
	switch strings.ToLower(args[0]) {
	case "head":
		hmethod = http.MethodHead
		HxHeaders = true
	case "get":
		hmethod = http.MethodGet
	case "post":
		hmethod = http.MethodPost
	case "put":
		hmethod = http.MethodPut
	case "patch":
		hmethod = http.MethodPatch
	case "del", "delete":
		hmethod = http.MethodDelete
	default:
		perr("ERROR invalid method [%s]", args[0])
		os.Exit(1)
	}

	if len(args) < 2 {
		perr("ERROR not enough arguments, see hx help/usage")
		os.Exit(1)
	}

	// https://pkg.go.dev/net/url#URL
	hurl, err := url.Parse(args[1])
	if err != nil {
		perr("ERROR invalid url [%s]", args[1])
		os.Exit(1)
	}

	perr("DEBUG hurl %#v", hurl)

	if hurl.Scheme == "" {
		hurl.Scheme = "http"
	}

	hquery := hurl.Query()
	if err != nil {
		perr("ERROR invalid url [%s] query part", hurl)
		os.Exit(1)
	}

	// https://pkg.go.dev/net/http#Header
	hheader := make(http.Header)

	for _, a := range args[2:] {
		// https://pkg.go.dev/strings#SplitN
		iheader := strings.Index(a, ":")
		iquery := strings.Index(a, "=")
		aisheader := iheader >= 0 && (iquery < 0 || iquery > iheader)
		aisquery := iquery >= 0 && (iheader < 0 || iheader > iquery)
		// https://pkg.go.dev/strings#SplitN
		if aisquery {
			akv := strings.SplitN(a, "=", 2)
			hquery.Add(akv[0], akv[1])
		} else if aisheader {
			akv := strings.SplitN(a, ":", 2)
			hheader.Add(akv[0], akv[1])
		} else {
			perr("ERROR invalid arg [%s]", a)
			os.Exit(1)
		}
	}

	hurl.RawQuery = hquery.Encode()

	perr("DEBUG hmethod [%s] hheader (%v) hurl %#v", hmethod, hheader, hurl)

	// https://pkg.go.dev/http#NewRequest
	hreq, err := http.NewRequest(hmethod, hurl.String(), nil)
	if err != nil {
		perr("ERROR NewRequest %v", err)
		os.Exit(1)
	}

	for hk, hvv := range hheader {
		for _, hv := range hvv {
			hreq.Header.Add(hk, hv)
		}
	}

	hclient := &http.Client{}
	if HxInsecure {
		hclient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	hresp, err := hclient.Do(hreq)
	if err != nil {
		perr("ERROR %v", err)
		os.Exit(1)
	}

	perr("DEBUG hresp %v", hresp)

	perr("DEBUG %s", hresp.Status)
	for hhk, hhvv := range hresp.Header {
		for _, hhv := range hhvv {
			perr("DEBUG %s: %s", hhk, hhv)
		}
	}

	defer hresp.Body.Close()

	if HxHeaders {
		pout(hresp.Status)
		for hhk, hhvv := range hresp.Header {
			for _, hhv := range hhvv {
				pout("%s: %s", hhk, hhv)
			}
		}
		pout("")
	}

	// https://pkg.go.dev/io#Copy
	_, err := io.Copy(os.Stdout, hresp.Body)
	if err != nil {
		perr("ERROR %v", err)
		os.Exit(1)
	}

	exitstatus := 0
	if hresp.StatusCode >= 400 {
		exitstatus = 1
	}

	os.Exit(exitstatus)
}

func tsnow() string {
	t := time.Now().Local()
	return fmt.Sprintf(
		"%03d:%02d%02d:%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
}

func fmtdur(t uint64) string {
	tdays, tsecs := t/(24*3600), t%(24*3600)
	durs := fmt.Sprintf("%ds", tsecs)
	if tdays > 0 {
		durs = fmt.Sprintf("%dd", tdays) + SEP + durs
	}
	return durs
}

func perr(msg string, args ...interface{}) {
	if strings.HasPrefix(msg, "DEBUG ") && !DEBUG {
		return
	}
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func pout(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stdout, msg+NL)
	} else {
		fmt.Fprintf(os.Stdout, msg+NL, args...)
	}
}

func seps(i uint64, e uint64) string {
	ee := uint64(math.Pow(10, float64(e)))
	if i < ee {
		return fmt.Sprintf("%d", i%ee)
	} else {
		f := fmt.Sprintf("0%dd", e)
		return fmt.Sprintf("%s"+SEP+"%"+f, seps(i/ee, e), i%ee)
	}
}
