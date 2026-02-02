/*
history:
2015/0428 v1
2021/0305 append mode
*/

/*
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
)

const (
	NL = "\n"
)

func log(msg string, args ...interface{}) {
	if len(os.Args) == 0 {
		fmt.Fprint(os.Stderr, msg+NL)
	} else {
		fmt.Fprintf(os.Stderr, msg+NL, args...)
	}
}

func main() {
	var err error
	var fpath string
	var mode os.FileMode = os.FileMode(0644)
	var modearg *os.FileMode

	if len(os.Args) <= 3 {
		fpath = os.Args[1]
	} else {
		log("usage: put path [mode]")
		os.Exit(1)
	}

	if len(os.Args) > 2 {
		var m uint64
		m, err = strconv.ParseUint(os.Args[2], 8, 32)
		if err != nil {
			log("ERROR invalid file mode [%s]", os.Args[2])
			os.Exit(1)
		}
		mode = os.FileMode(m)
		modearg = &mode
	}

	dirpath := path.Dir(fpath)
	dirstat, err := os.Stat(dirpath)
	if err == nil && !dirstat.IsDir() {
		log("ERROR [%s] is not a dir", dirpath)
		os.Exit(1)
	}
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirpath, os.FileMode(0755))
		if err != nil {
			log("ERROR %v", err)
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
		log("ERROR %v", err)
		os.Exit(1)
	}
	defer f.Close()

	if modearg != nil {
		if err := f.Chmod(mode); err != nil {
			log("ERROR %v", err)
			os.Exit(1)
		}
	}

	if truncatefile {
		if err := f.Truncate(0); err != nil {
			log("ERROR %v", err)
			os.Exit(1)
		}
	}

	if _, err := io.Copy(f, os.Stdin); err != nil {
		log("ERROR %v", err)
		os.Exit(1)
	}

	if err := f.Sync(); err != nil {
		log("ERROR %v", err)
		os.Exit(1)
	}

	if err := f.Close(); err != nil {
		log("ERROR %v", err)
		os.Exit(1)
	}
}
