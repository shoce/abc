/*
history:
2015-04-19 v1
2020-0127 ignore SIGURG
2025-0807 seps + end of stdin
*/
/*
GoGet
GoBuild
*/
/*
pipemon </dev/random >/dev/null
pipemon </etc/passwd >/dev/null
pipemon </dev/null >/dev/null
*/

package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	SEP = ","
	NL = "\n"
	CopyNBytes = 64 << 10
)

var (
	err error
	t0 time.Time
	written uint64
	
	F = fmt.Sprintf
	FI = strconv.FormatInt
)

func copy(ch chan error) {
	var w int64
	for err == nil {
		// https://pkg.go.dev/io#CopyN
		w, err = io.CopyN(os.Stdout, os.Stdin, CopyNBytes)
		written = written + uint64(w)
	}
	ch <- err
}

func main() {
	t0 = time.Now()
	var sigchan = make(chan os.Signal)
	signal.Notify(sigchan)
	var copychan = make(chan error)
	go copy(copychan)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			report()
		}
	}()
	for {
		select {
		case s := <-sigchan:
			if s == syscall.SIGURG { continue }
			perr(F("signal %v", s))
			report()
			os.Exit(1)
		case e := <-copychan:
			if e == io.EOF {
				report()
				perr("end of stdin")
				os.Exit(0)
			} else {
				perr(F("error copy %v", e))
				report()
				os.Exit(1)
			}
		}
	}
}

func report() {
	dt := time.Since(t0).Seconds()
	perr(F("time <%s,s> written <%s,kb> rate <%s,kbps>",
		seps(int(dt), 2),
		seps(int(written>>10), 3),
		seps(int(float64(written>>10)/dt), 3),
	))
}

func seps(i, e int) string {
	ee := int(math.Pow(10, float64(e)))
	if i < ee { return FI(int64(i%ee), 10) }
	f := "%0"+FI(int64(e), 10)+"d"
	return seps(i/ee, e)+SEP+F(f , i%ee)
}

func perr(msg string) { fmt.Fprint(os.Stderr, "pipemon "+msg+NL) }


