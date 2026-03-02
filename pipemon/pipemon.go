/*
history:
2015-04-19 v1
2020-0127 ignore SIGURG
2025-0807 seps + end of stdin
*/

// GoGet GoFmt GoBuild

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
	"syscall"
	"time"
)

const (
	SEP = ","

	NL = "\n"

	CopyNBytes = 64 << 10
)

var (
	err     error
	t0      time.Time
	written uint64
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
			if s == syscall.SIGURG {
				continue
			}
			perr("signal %v", s)
			report()
			os.Exit(1)
		case e := <-copychan:
			if e == io.EOF {
				report()
				perr("end of stdin")
				os.Exit(0)
			} else {
				perr("error copy %v", e)
				report()
				os.Exit(1)
			}
		}
	}
}

func seps(i uint64, e int) string {
	ee := uint64(math.Pow(10, float64(e)))
	if i < ee {
		return fmt.Sprintf("%d", i%ee)
	} else {
		f := fmt.Sprintf("0%dd", e)
		return fmt.Sprintf("%s"+SEP+"%"+f, seps(i/ee, e), i%ee)
	}
}

func perr(msg string, args ...interface{}) {
	msgtext := msg
	if len(args) > 0 {
		msgtext = fmt.Sprintf(msg, args...)
	}
	fmt.Fprint(os.Stderr, "pipemon "+msgtext+NL)
}

func report() {
	dt := time.Since(t0).Seconds()
	perr("time <%ss> written <%skb> rate <%skbps>",
		seps(uint64(dt), 2),
		seps(written>>10, 3),
		seps(uint64(float64(written>>10)/dt), 3),
	)
}
