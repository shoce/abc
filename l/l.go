/*
history:
016/0608 v1
020/1106 print only file name by default, add options -r, -s, -t, -m, -l
021/0329 add cid printing
021/1026 add -1 option

GoGet GoFmt GoBuildNull
GoBuild
GoRun

 && ln -sf l /bin/ls && ln -sf l /bin/lsr && ln -sf l /bin/lt && ln -sf l /bin/ll && ln -sf l /bin/lr && ln -sf l /bin/llr

*/

package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
)

const (
	NL  = "\n"
	TAB = "\t"
)

var (
	VERSION string

	TERM string

	Recursive   bool
	ShowSymlink bool
	ShowTime    bool
	ShowSize    bool
	ShowPerm    bool
	ShowOwner   bool
	ShowCid     bool
)

func printinfo(path string, info os.FileInfo) error {
	var err error

	s := path

	s = strings.ReplaceAll(s, TAB, "\\\t")

	var finfo os.FileInfo = info

	if finfo.Mode().IsDir() {
		s += string(os.PathSeparator)
	}

	if fperm := finfo.Mode().Perm(); fperm&0111 != 0 {
		//s = TermBold(s)
		s = TermItalic(s)
		//s = TermUnderline(s)
	}

	if ShowSymlink && (finfo.Mode()&os.ModeSymlink) != 0 {
		finfo, err = os.Lstat(path)
		if err != nil {
			return err
		}
		var linkpath string
		linkpath, err = os.Readlink(path)
		if err != nil {
			return err
		}
		s += TAB + fmt.Sprintf("symlink[%s]", linkpath)
	}

	if ShowPerm {
		s += TAB + fmt.Sprintf("perm:%04o", finfo.Mode().Perm())
	}

	if ShowOwner {
		var fstatuid, fstatgid int64 = -1, -1
		if fstat, ok := finfo.Sys().(*syscall.Stat_t); ok {
			fstatuid, fstatgid = int64(fstat.Uid), int64(fstat.Gid)
		}
		s += TAB + fmt.Sprintf("uid:%d gid:%d", fstatuid, fstatgid)
	}

	if ShowSize && !finfo.Mode().IsDir() && (info.Mode()&os.ModeSymlink == 0) {
		s += TAB + fmt.Sprintf("size:%s", seps(int(finfo.Size()), 3))
	}
	if ShowSize && finfo.Mode().IsDir() {
		s += TAB + "size:dir"
	}
	if ShowSize && (info.Mode()&os.ModeSymlink != 0) {
		s += TAB + "size:symlink"
	}

	if ShowTime {
		s += TAB + fmt.Sprintf("mtime:%s", finfo.ModTime().UTC().Format("06.0102.1504"))
	}

	if ShowCid && !finfo.IsDir() && (info.Mode()&os.ModeSymlink == 0) {
		f, err := os.Open(path)
		if err != nil {
			perr("%v", err)
			return err
		}
		defer f.Close()
		fmh, err := mh.SumStream(f, mh.SHA2_256, -1)
		if err != nil {
			perr("%v", err)
			return err
		}
		c := cid.NewCidV1(cid.Raw, fmh)
		s += TAB + fmt.Sprintf("cid:%s", c)
	}

	fmt.Println(s)
	return nil
}

func fls(path string, info os.FileInfo, err error) error {
	if err != nil {
		perr("%v", err)
		return err
	}
	if err2 := printinfo(path, info); err2 != nil {
		perr("%v", err2)
		return err2
	}
	return nil
}

func list(path string) error {
	var listdir bool
	if strings.HasSuffix(path, "/") {
		listdir = true
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pathstat, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if (pathstat.Mode() & os.ModeSymlink) != 0 {
		linktargetpath, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(linktargetpath) {
			linktargetpath = filepath.Clean(filepath.Join(filepath.Dir(path), linktargetpath))
		}
		linktargetstat, err := os.Lstat(linktargetpath)
		if err != nil {
			return err
		}
		if !linktargetstat.Mode().IsDir() && listdir {
			return fmt.Errorf("%s symblink[%s] is not a dir", path, linktargetpath)
		}
	} else if !pathstat.Mode().IsDir() && listdir {
		return fmt.Errorf("%s is not a dir", path)
	}

	if Recursive {
		err = filepath.Walk(path, fls)
		if err != nil {
			return err
		}
	} else if listdir {
		if err := printinfo(path, pathstat); err != nil {
			return err
		}

		ff, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		for _, fstat := range ff {
			fpath := filepath.Join(path, fstat.Name())
			printinfo(fpath, fstat)
		}
	} else {
		if err := printinfo(path, pathstat); err != nil {
			return err
		}
	}

	return nil
}
func init() {
	if len(os.Args) == 2 && os.Args[1] == "-version" {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if v := os.Getenv("TERM"); v != "" {
		TERM = v
	}
}

func main() {
	var err error

	cmdname := path.Base(os.Args[0])
	switch cmdname {
	case "ls":
		ShowSize = true
		ShowSymlink = true
	case "lsr":
		Recursive = true
		ShowSize = true
		ShowSymlink = true
	case "lt":
		ShowTime = true
	case "lr":
		Recursive = true
	case "ll":
		ShowSymlink = true
		ShowPerm = true
		ShowOwner = true
		//ShowTime = true
		ShowSize = true
		//ShowCid = true
	case "llr":
		Recursive = true
		ShowSymlink = true
		ShowPerm = true
		ShowOwner = true
		//ShowTime = true
		ShowSize = true
		//ShowCid = true
	}

	paths := os.Args[1:]
	for _, p := range paths {
		if !strings.HasPrefix(p, "-") {
			break
		}
		switch p {
		case "-r", "-recursive":
			Recursive = true
		case "-p", "-perm":
			ShowPerm = true
		case "-o", "-owner":
			ShowOwner = true
		case "-s", "-size":
			ShowSize = true
		case "-t", "-time":
			ShowTime = true
		case "-c", "-cid":
			ShowCid = true
		case "-l", "-link":
			ShowSymlink = true
		case "-1", "-pathonly":
			ShowPerm = false
			ShowSymlink = false
			ShowTime = false
			ShowSize = false
			ShowCid = false
			ShowOwner = false
		default:
			perr("invalid option [%s]", p)
			os.Exit(1)
		}
		paths = paths[1:]
	}

	if len(paths) == 0 {
		paths = append(paths, "./")
	}

	for _, p := range paths {
		err = list(p)
		if err != nil {
			perr("%v", err)
			os.Exit(1)
		}
	}
}

func TermBold(s string) string {
	if TERM != "" {
		return "\033[1m" + s + "\033[21m"
	}
	return s
}

func TermItalic(s string) string {
	if TERM != "" {
		return "\033[3m" + s + "\033[23m"
	}
	return s
}

func TermUnderline(s string) string {
	if TERM != "" {
		return "\033[4m" + s + "\033[24m"
	}
	return s
}

func seps(i, e int) string {
	ee := int(math.Pow(10, float64(e)))
	if i < ee {
		return fmt.Sprintf("%d", i%ee)
	} else {
		f := fmt.Sprintf("0%dd", e)
		return fmt.Sprintf("%s.%"+f, seps(i/ee, e), i%ee)
	}
}

func ts() string {
	t := time.Now().UTC()
	return fmt.Sprintf(
		"%03d:%02d%02d:%02d%02d+",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
}

func perr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, ts()+" "+msg+NL, args...)
}
