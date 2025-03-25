/*

go get -u -v
go mod tidy

GoFmt
GoBuildNull
GoBuild
GoRun

*/

package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

const (
	NL = "\n"

	ShowLoopback = false
	ShowDown     = false
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

	netinterfaces := make(map[string]NetInterface)

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
			ni.Addr = append(ni.Addr, a.String())
		}

		netinterfaces[ni.Name] = ni
	}

	ye := yaml.NewEncoder(os.Stdout)
	defer ye.Close()
	err = ye.Encode(netinterfaces)
	if err != nil {
		fmt.Fprintf(os.Stderr, "yaml.Encoder.Encode: %v"+NL, err)
	}
}

type NetInterface struct {
	Name         string   `yaml:"name"`
	Loopback     bool     `yaml:"loopback"`
	PointToPoint bool     `yaml:"ptp"`
	Up           bool     `yaml:"up"`
	HwAddr       string   `yaml:"hwaddr"`
	Addr         []string `yaml:"addr"`
	Error        error    `yaml:"error,omitempty"`
}
