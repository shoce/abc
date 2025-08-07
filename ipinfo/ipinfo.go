/*
GoGet GoFmt GoBuildNull
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
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: ipinfo <ip-address>"+NL)
		os.Exit(1)
	}

	ipaddr := os.Args[1]
	requrl := fmt.Sprintf("https://ipinfo.io/%s/json", ipaddr)

	resp, err := http.Get(requrl)
	if err != nil {
		fmt.Println("error http.Get:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	var ipinfo IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&ipinfo); err != nil {
		fmt.Println("error json.Decode:", err)
		os.Exit(1)
	}

	fmt.Printf(
		"ip <%s> country [%s] region [%s] org [%s]"+NL,
		ipaddr, ipinfo.Country, ipinfo.Region, ipinfo.Org,
	)
}
