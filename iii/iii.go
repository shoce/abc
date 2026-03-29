// GoFixDiff GoGet GoFmt GoBuildNull
// GoBuild GoRun

package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	SP  = " "
	TAB = "\t"
	NL  = "\n"
)

var (
	VERSION string

	ShowEmptyAddr = false
	ShowDown      = false
	ShowLoopback  = false
	ShowPtp       = true

	III map[string]NetInterface
)

func main() {
	var err error
	var ii []net.Interface
	var i net.Interface
	var aa []net.Addr
	var a net.Addr

	var flags, args []string
	for _, arg := range os.Args[1:] {
		if arg == "" {
			continue
		}
		if arg[0] == '-' {
			flags = append(flags, arg)
		} else {
			args = append(args, arg)
		}
	}
	//fmt.Printf("flags %#v args %#v"+NL, flags, args)
	for _, f := range flags {
		switch f {
		case "-version":
			fmt.Print(VERSION + NL)
			os.Exit(0)
		case "-a":
			ShowEmptyAddr = true
			ShowDown = true
			ShowLoopback = true
			ShowPtp = true
		}
	}

	// https://pkg.go.dev/net#Interfaces
	ii, err = net.Interfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR net.Interfaces() %v"+NL, err)
		os.Exit(1)
	}

	III = make(map[string]NetInterface)

	for _, i = range ii {
		if len(args) > 0 {
			listit := false
			for _, a := range args {
				if strings.HasPrefix(i.Name, a) {
					listit = true
				}
			}
			if !listit {
				continue
			}
		}

		ni := NetInterface{Name: i.Name, HwAddr: fmt.Sprintf("%v", i.HardwareAddr)}

		if i.Flags&net.FlagLoopback != 0 {
			ni.Loopback = true
		}
		if i.Flags&net.FlagUp != 0 {
			ni.Up = true
		}
		if i.Flags&net.FlagPointToPoint != 0 {
			ni.PointToPoint = true
		}

		if !ShowDown && !ni.Up {
			continue
		}
		if !ShowLoopback && ni.Loopback {
			continue
		}
		if !ShowPtp && ni.PointToPoint {
			continue
		}

		// https://pkg.go.dev/net#Interface.Addrs
		aa, err = i.Addrs()
		if err != nil {
			ni.Error = fmt.Sprintf("%s", err)
			continue
		}
		// https://pkg.go.dev/slices#SortFunc
		slices.SortFunc(aa, func(a, b net.Addr) int {
			// https://pkg.go.dev/net#ParseCIDR
			ipa, _, _ := net.ParseCIDR(a.String())
			ipb, _, _ := net.ParseCIDR(b.String())
			return bytes.Compare(ipa, ipb)
		})
		for _, a = range aa {
			ni.Addr = append(ni.Addr, fmt.Sprintf("<%s>", a))
		}

		III[ni.Name] = ni
	}

	for _, ni := range III {
		if !ShowEmptyAddr && len(ni.Addr) == 0 {
			continue
		}
		iinfo := fmt.Sprintf(
			"[%s]"+SP+"addr( %s )"+SP+"hwaddr<%s>",
			ni.Name, strings.Join(ni.Addr, SP), ni.HwAddr,
		)
		if ShowDown {
			iinfo += SP + fmt.Sprintf("up<%t>", ni.Up)
		}
		if ShowLoopback {
			iinfo += SP + fmt.Sprintf("loopback<%t>", ni.Loopback)
		}
		if ShowPtp {
			iinfo += SP + fmt.Sprintf("ptp<%t>", ni.PointToPoint)
		}
		if ni.Error != "" {
			iinfo += SP + fmt.Sprintf("err[%s]", ni.Error)
		}
		fmt.Print(iinfo + NL)
	}
}

type NetInterface struct {
	Name         string
	Addr         []string
	HwAddr       string
	Up           bool
	Loopback     bool
	PointToPoint bool
	Error        string
}
