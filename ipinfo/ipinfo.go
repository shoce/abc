/*
GoGet
GoFmt
GoBuildNull
GoBuild
*/
package main
import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)
const (
	NL = "\n"
)
var (
	F = fmt.Sprintf
	pout = fmt.Print
)
func perr(s string) (int, error) { return fmt.Fprint(os.Stderr, s) }
type IPInfo struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Loc      string `json:"loc"`
	Org      string `json:"org"`
	Postal   string `json:"postal"`
}
func main() {
	if len(os.Args)>2 {
		perr("usage: ipinfo [<ip-address>...]"+NL)
		os.Exit(1)		
	}
	aa := []string{""}
	if len(os.Args)>1 { aa = os.Args[1:] }
	for _, ipaddr := range aa {
		requrl := F("https://ipinfo.io/%s/json", ipaddr)
		perr(F("DEBUG requrl [%s]", requrl)+NL)
		// https://pkg.go.dev/net/http#NewRequest
		req, err := http.NewRequest(http.MethodGet, requrl, nil)
		if err!=nil { perr(F("ERROR http.NewRequest [%s] %v", requrl, err)+NL) }
		req.Header.Set("Accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			perr(F("ERROR http.Get [%s] %v", requrl, err)+NL)
			continue
		}
		defer resp.Body.Close()
		var ipinfo IPInfo
		if err := json.NewDecoder(resp.Body).Decode(&ipinfo); err != nil {
			perr(F("ERROR json.Decode %v", err)+NL)
			continue
		}
		pout(F(
			"@ip <%s> @country [%s] @region [%s] @org [%s]"+NL,
			ipinfo.IP, ipinfo.Country, ipinfo.Region, ipinfo.Org,
		))
	}
}
