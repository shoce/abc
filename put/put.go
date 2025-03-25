/*
history:
2015/0428 v1
2021/0305 append mode

usage:
put put-file-test 600 <some-another-file
id | put id.out.text
sudo id | sudo put sudo.id.out.text

GoFmt GoBuildNull GoBuild

curl -sSL https://github.com/shoce/put/releases/latest/download/put.linux.gz | gunzip >/bin/put && chmod 755 /bin/put && ln -sf put /bin/append
*/

package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"
)

func log(msg string, args ...interface{}) {
	const Beat = time.Duration(24) * time.Hour / 1000
	tzBiel := time.FixedZone("Biel", 60*60)
	t := time.Now().In(tzBiel)
	ty := t.Sub(time.Date(t.Year(), 1, 1, 0, 0, 0, 0, tzBiel))
	td := t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tzBiel))
	ts := fmt.Sprintf(
		"%d/%d@%d",
		t.Year()%1000,
		int(ty/(time.Duration(24)*time.Hour))+1,
		int(td/Beat),
	)
	fmt.Fprintf(os.Stderr, ts+" "+msg+"\n", args...)
}

func main() {
	var err error
	var fpath string
	var mode os.FileMode = os.FileMode(0644)
	var modearg *os.FileMode

	if len(os.Args) > 1 {
		fpath = os.Args[1]
	} else {
		log("usage: put path [mode]")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		var m uint64
		m, err = strconv.ParseUint(os.Args[2], 8, 32)
		if err != nil {
			log("invalid file mode `%s`", os.Args[2])
			os.Exit(1)
		}
		mode = os.FileMode(m)
		modearg = &mode
	}

	dirpath := path.Dir(fpath)
	dirstat, err := os.Stat(dirpath)
	if err == nil && !dirstat.IsDir() {
		log("%s is not a dir", dirpath)
		os.Exit(1)
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirpath, os.FileMode(0755))
		if err != nil {
			log("%v", err)
			os.Exit(1)
		}
	}

	var truncatefile bool
	if path.Base(os.Args[0]) == "put" {
		truncatefile = true
	}

	fflag := os.O_CREATE | os.O_WRONLY
	if path.Base(os.Args[0]) == "append" {
		fflag |= os.O_APPEND
	}

	var f *os.File
	f, err = os.OpenFile(fpath, fflag, mode)
	if err != nil {
		log("%v", err)
		os.Exit(1)
	}
	defer f.Close()

	if modearg != nil {
		if err := f.Chmod(mode); err != nil {
			log("%v", err)
			os.Exit(1)
		}
	}

	if truncatefile {
		if err := f.Truncate(0); err != nil {
			log("%v", err)
			os.Exit(1)
		}
	}

	if _, err := io.Copy(f, os.Stdin); err != nil {
		log("%v", err)
		os.Exit(1)
	}

	if err := f.Sync(); err != nil {
		log("%v", err)
		os.Exit(1)
	}

	if err := f.Close(); err != nil {
		log("%v", err)
		os.Exit(1)
	}
}
