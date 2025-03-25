/*
history:
2015-04-19 v1
2020-0127 ignore SIGURG

GoFmt GoBuild GoRelease

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

const N = 64 * 1024

var (
	err     error
	t0      time.Time
	written int64
)

func errprintf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, "pipemon: "+format, a...)
}

func report() {
	var dt float64
	dt = time.Since(t0).Seconds()
	var kbps int64
	kbps = int64(float64(written/1024) / dt)
	errprintf("time=%ds written=%dkb rate=%dkbps\r\n", int64(dt), written/1024, kbps)
}

func copy(ch chan error) {
	var w int64
	for err == nil {
		w, err = io.CopyN(os.Stdout, os.Stdin, N)
		written = written + w
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
			errprintf("signal: %v\n", s)
			report()
			os.Exit(1)
		case e := <-copychan:
			if e == io.EOF {
				report()
				os.Exit(0)
			} else {
				errprintf("copy: %v\n", e)
				report()
				os.Exit(1)
			}
		}
	}
}
