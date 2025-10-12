/*

GoGet GoFmt GoBuildNull
GoBuild GoRun

*/

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	SP  = " "
	TAB = "\t"
	NL  = "\n"

	ShowLoopback = false
	ShowDown     = false
)

var (
	NETINTERFACES map[string]NetInterface
)

func main() {
	var err error
	var ii []net.Interface
	var i net.Interface
	var aa []net.Addr
	var a net.Addr

	ii, err = net.Interfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "net.Interfaces(): %v"+NL, err)
	}

	NETINTERFACES = make(map[string]NetInterface)

	for _, i = range ii {
		if len(os.Args) > 1 {
			listit := false
			for _, a := range os.Args {
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

		if !ShowLoopback && ni.Loopback {
			continue
		}
		if !ShowDown && !ni.Up {
			continue
		}

		aa, err = i.Addrs()
		if err != nil {
			ni.Error = err
			continue
		}

		for _, a = range aa {
			ni.Addr = append(ni.Addr, fmt.Sprintf("<%s>", a))
		}

		NETINTERFACES[ni.Name] = ni
	}

	for _, ni := range NETINTERFACES {
		fmt.Printf(
			"@%s {"+NL+
				TAB+"@addr ( %s )"+NL+
				TAB+"@hwaddr <%s>"+NL+
				TAB+"@up <%t>"+NL+
				TAB+"@loopback <%t>"+NL+
				TAB+"@ptp <%t>"+NL+
				TAB+"@err %v"+NL+
				"}"+NL,
			ni.Name, strings.Join(ni.Addr, SP), ni.HwAddr, ni.Up, ni.Loopback, ni.PointToPoint, ni.Error,
		)
	}
}

type NetInterface struct {
	Name         string
	Addr         []string
	HwAddr       string
	Up           bool
	Loopback     bool
	PointToPoint bool
	Error        error
}
