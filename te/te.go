/*
history:
2020/04/17 v1

https://pkg.go.dev/text/template

go get -u -v
go mod tidy

GoFmt
GoBuildNull
GoBuild
*/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
	"time"
)

var (
	te   *template.Template
	tefm template.FuncMap
)

func init() {
	tefm = template.FuncMap{
		"Files":       Files,
		"Dirs":        Dirs,
		"FilesDirs":   FilesDirs,
		"DirsFiles":   DirsFiles,
		"DirName":     DirName,
		"ReadFile":    ReadFile,
		"UserHomeDir": os.UserHomeDir,
		"HasPrefix":   strings.HasPrefix,
		"HasSuffix":   strings.HasSuffix,
		"TrimPrefix":  strings.TrimPrefix,
		"TrimSuffix":  strings.TrimSuffix,
		"TrimSpace":   strings.TrimSpace,
		"Contains":    strings.Contains,
		"Join":        strings.Join,
		"Split":       strings.Split,
		"Index":       Index,
		"Append":      Append,
		"MapNew":      MapNew,
		"MapAppend":   MapAppend,
	}

	te = template.New("te")
	te = te.Funcs(tefm)
}

func main() {
	var err error
	var inbb, outbb []byte
	var in string
	var inpath, outpath string

	if len(os.Args) > 1 {
		inpath = os.Args[1]
	}
	if len(os.Args) > 2 {
		outpath = os.Args[2]
	}

	if inpath != "" {
		inbb, err = ioutil.ReadFile(inpath)
		if err != nil {
			log("read %s: %s", inpath, err)
			os.Exit(1)
		}
	} else {
		inbb, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			log("read stdin: %s", err)
			os.Exit(1)
		}
	}
	in = string(inbb)

	te, err = te.Parse(in)
	if err != nil {
		log("parse: %s", err)
		os.Exit(1)
	}

	if outpath != "" {
		bb := bytes.NewBuffer(nil)
		err = te.Execute(bb, nil)
		if err != nil {
			log("execute: %s", err)
			os.Exit(1)
		}

		outbb, err = ioutil.ReadFile(outpath)
		if err == nil {
			if bytes.Equal(bb.Bytes(), outbb) {
				//log("same output as %s, not writing", outpath)
				os.Exit(0)
			}
		}

		err = ioutil.WriteFile(outpath, bb.Bytes(), 0644)
		if err != nil {
			log("write %s: %s", outpath, err)
			os.Exit(1)
		}
		log("wrote %d bytes to %s", len(bb.Bytes()), outpath)
	} else {
		err = te.Execute(os.Stdout, nil)
		if err != nil {
			log("execute: %s", err)
			os.Exit(1)
		}
	}
}

func log(msg string, args ...interface{}) {
	ts := time.Now().Local().Format("Jan/02;15:04")
	fmt.Fprintf(os.Stderr, ts+" "+msg+"\n", args...)
}

func Files(fp string) ([]string, error) {
	var err error

	fi, err := os.Lstat(fp)
	if err != nil {
		return nil, err
	}

	if !fi.Mode().IsDir() {
		return nil, nil
	}
	ff, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, i := range ff {
		if i.Mode()&os.ModeType != 0 {
			continue
		}
		if strings.HasPrefix(i.Name(), ".") {
			continue
		}
		names = append(names, i.Name())
	}

	return names, nil
}

func Dirs(fp string) ([]string, error) {
	var err error

	fi, err := os.Lstat(fp)
	if err != nil {
		return nil, err
	}

	if !fi.Mode().IsDir() {
		return nil, nil
	}
	ff, err := ioutil.ReadDir(fp)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, i := range ff {
		if !i.Mode().IsDir() {
			continue
		}
		if strings.HasPrefix(i.Name(), ".") {
			continue
		}
		names = append(names, i.Name())
	}

	return names, nil
}

func FilesDirs(fp string) ([]string, error) {
	ff, err := Files(fp)
	if err != nil {
		return nil, err
	}
	dd, err := Dirs(fp)
	if err != nil {
		return nil, err
	}
	rr := append(ff, dd...)
	return rr, nil
}

func DirsFiles(fp string) ([]string, error) {
	dd, err := Dirs(fp)
	if err != nil {
		return nil, err
	}
	ff, err := Files(fp)
	if err != nil {
		return nil, err
	}
	rr := append(dd, ff...)
	return rr, nil
}

func DirName() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	return path.Base(wd), nil
}

func ReadFile(fp string) (string, error) {
	bb, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", nil
		//return "", fmt.Errorf("read %s: %s", fp, err)
	}
	return string(bb), nil
}

func Index(a []string, i int) (string, error) {
	if i == 0 {
		return "", fmt.Errorf("index 0 is invalid", i)
	}
	if i > len(a) {
		return "", fmt.Errorf("index %d is out of range, length is %d", i, len(a))
	}
	if i < 0 {
		return a[len(a)+i], nil
	}
	return a[i-1], nil
}

func Append(a []string, b string) ([]string, error) {
	return append(a, b), nil
}

func MapNew() (map[string][]string, error) {
	m := make(map[string][]string)
	return m, nil
}

func MapAppend(m map[string][]string, a, b string) (map[string][]string, error) {
	if _, ok := m[a]; !ok {
		m[a] = []string{}
	}
	m[a] = append(m[a], b)
	return m, nil
}
