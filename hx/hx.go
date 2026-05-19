/*
HISTORY
26/0330@thailand v1
026/0519 func args()
*/

// GoGet GoFmt GoBuildNull
// GoBuild GoRun

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptrace"
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
	USAGE   string = `USAGE
hx get scheme://host:port/path/subpath header1:v1 header2:v2 arg1=val1 arg2=val2
`
	DEBUG bool

	USERAGENT = "hx/1.0"

	HxHeaders bool

	// TODO basic auth
	HxUser string
	HxPass string

	// TODO timeout
	HxTimeout time.Duration

	HxInsecure bool

	F = fmt.Sprintf
	pout = fmt.Print
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

func argss() (args []string) {
args = os.Args[1:]

fd3 := os.NewFile(3, "fd3")
if fd3 == nil { perr("ERROR NewFile <3>"); return; }
defer fd3.Close()
data, _ := io.ReadAll(fd3)
if len(data) == 0 { return; }
if len(data)==1 && data[0]=='\n' { return; }

// https://pkg.go.dev/strings#Split
args = strings.Split(string(data), NL)
if len(args)>0 && args[len(args)-1]=="" { 
args = args[:len(args)-1] 
}
return

}

func main() {
	var err error

	args := argss()
	perr(F("DEBUG args %#v", args))
	n := 0
	for _, a := range args {
		if a != "" {
			args[n] = a
			n++
		}
	}
	args = args[:n]
	perr(F("DEBUG n <%d> args %#v", n, args))

	if len(args) == 1 && args[0] == "version" {
		pout(VERSION + NL)
		os.Exit(0)
	}

	if len(args) == 1 && slices.Contains([]string{"help", "usage"}, args[0]) {
		pout(USAGE + NL)
		os.Exit(0)
	}

	if len(args) < 1 {
		perr("ERROR not enough arguments")
		perr(USAGE + NL)
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
		perr(F("ERROR invalid method [%s]", args[0]))
		os.Exit(1)
	}

	if len(args) < 2 {
		perr("ERROR not enough arguments")
		perr(USAGE + NL)
		os.Exit(1)
	}

	// https://pkg.go.dev/net/url#URL
	hurl, err := url.Parse(args[1])
	if err != nil {
		perr(F("ERROR invalid url [%s]", args[1]))
		os.Exit(1)
	}

	perr(F("DEBUG hurl %#v", hurl))

	if hurl.Scheme == "" {
		hurl.Scheme = "http"
	}

	hquery := hurl.Query()
	if err != nil {
		perr(F("ERROR invalid url [%s] query part", hurl))
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
			perr(F("ERROR invalid arg [%s]", a))
			os.Exit(1)
		}
	}

	hurl.RawQuery = hquery.Encode()

	perr(F("DEBUG hmethod [%s]", hmethod))
	hheaderss := make([]string, 0)
	// https://pkg.go.dev/http#Request.Header
	for hk, hvv := range hheader {
		for _, hv := range hvv {
			hheaderss = append(hheaderss, F("%s[%s]", hk, hv))
		}
	}
	slices.SortFunc(hheaderss, strings.Compare)
	for _, h := range hheaderss {
		perr(F("DEBUG hheader %s", h))
	}
	perr(F("DEBUG hurl %#v", hurl))

	// https://pkg.go.dev/http#Client
	hclient := &http.Client{}
	if HxInsecure {
		hclient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	htrace := &httptrace.ClientTrace{
		DNSStart:     func(info httptrace.DNSStartInfo) {},
		DNSDone:      func(info httptrace.DNSDoneInfo) {},
		ConnectStart: func(network, addr string) {},
		GotConn:      func(info httptrace.GotConnInfo) {},
		TLSHandshakeStart: func() {
		},
		TLSHandshakeDone: func(htlsconnstate tls.ConnectionState, err error) {
			perr(F("DEBUG tls connection state %#v", htlsconnstate))
			for _, pc := range htlsconnstate.PeerCertificates {
				perr(F(
					"DEBUG tls connection peer certificate Issuer [%v] Subject [%v] NotBefore <%s> NotAfter <%s> KeyUsage [%v]",
					pc.Issuer, pc.Subject, fmttime(pc.NotBefore), fmttime(pc.NotAfter), pc.KeyUsage,
				))
				perr(F(
					"DEBUG tls connection peer certificate PermittedDNSDomains (%v) PermittedIPRanges (%v) PermittedEmailAddresses (%v) PermittedURIDomains (%v)",
					pc.PermittedDNSDomains, pc.PermittedIPRanges, pc.PermittedEmailAddresses, pc.PermittedURIDomains,
				))
			}
		},
		GotFirstResponseByte: func() {},
	}

	// https://pkg.go.dev/http#NewRequest
	hreq, err := http.NewRequest(hmethod, hurl.String(), nil)
	if err != nil {
		perr(F("ERROR NewRequest %v", err))
		os.Exit(1)
	}

	hreq = hreq.WithContext(httptrace.WithClientTrace(context.Background(), htrace))

	hreq.Header.Set("User-Agent", USERAGENT)

	for hk, hvv := range hheader {
		for _, hv := range hvv {
			hreq.Header.Add(hk, hv)
		}
	}

	hheaderss = make([]string, 0)
	// https://pkg.go.dev/http#Request.Header
	for hk, hvv := range hreq.Header {
		for _, hv := range hvv {
			hheaderss = append(hheaderss, F("%s[%s]", hk, hv))
		}
	}
	slices.SortFunc(hheaderss, strings.Compare)
	for _, h := range hheaderss {
		perr(F("DEBUG hreq.Header %s", h))
	}

	hresp, err := hclient.Do(hreq)
	if err != nil {
		perr(F("ERROR %v", err))
		os.Exit(1)
	}

	perr(F("DEBUG hresp %v", hresp))

	perr(F("DEBUG %s", hresp.Status))
	for hhk, hhvv := range hresp.Header {
		for _, hhv := range hhvv {
			perr(F("DEBUG %s: %s", hhk, hhv))
		}
	}

	defer hresp.Body.Close()

	if HxHeaders {
		pout(hresp.Status+NL)
		for hhk, hhvv := range hresp.Header {
			for _, hhv := range hhvv {
				pout(F("%s [%s]", hhk, hhv)+NL)
			}
		}
		pout(NL)
	}

	// https://pkg.go.dev/io#Copy
	_, err = io.Copy(os.Stdout, hresp.Body)
	if err != nil {
		perr(F("ERROR copy response body %v", err))
		os.Exit(1)
	}

	exitstatus := 0
	if hresp.StatusCode >= 400 {
		exitstatus = 1
	}

	os.Exit(exitstatus)
}

func fmttime(t time.Time) string {
	return F(
		"%d:%02d%02d:%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
}

func fmtdur(t uint64) string {
	tdays, tsecs := t/(24*3600), t%(24*3600)
	durs := F("%ds", tsecs)
	if tdays > 0 {
		durs = F("%dd", tdays) + SEP + durs
	}
	return durs
}

func perr(msgtext string) {
	if strings.HasPrefix(msgtext, "DEBUG ") && !DEBUG {
		return
	}
	fmt.Fprint(os.Stderr, msgtext+NL)
}

func seps(i uint64, e uint64) string {
	ee := uint64(math.Pow(10, float64(e)))
	if i < ee {
		return F("%d", i%ee)
	} else {
		f := F("0%dd", e)
		return F("%s"+SEP+"%"+f, seps(i/ee, e), i%ee)
	}
}
