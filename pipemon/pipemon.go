/*
history:
2015-04-19 v1
2020-0127 ignore SIGURG

GoGet GoFmt GoBuild

pipemon </dev/random >/dev/null
pipemon </etc/passwd >/dev/null
pipemon </dev/null >/dev/null
*/

package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	NL = "\n"

	N = 64 << 10
)

var (
	err     error
	t0      time.Time
	written uint64
)

func log(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, "pipemon: "+format+NL, a...)
}

func report() {
	dt := time.Since(t0).Seconds()
	log("time <%ds> written <%dkb> rate <%dkbps>",
		uint64(dt),
		written>>10,
		uint64(float64(written>>10)/dt),
	)
}

func copy(ch chan error) {
	var w int64
	for err == nil {
		w, err = io.CopyN(os.Stdout, os.Stdin, N)
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
			log("signal: %v", s)
			report()
			os.Exit(1)
		case e := <-copychan:
			if e == io.EOF {
				report()
				os.Exit(0)
			} else {
				log("copy: %v", e)
				report()
				os.Exit(1)
			}
		}
	}
}
