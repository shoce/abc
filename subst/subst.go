/*
history:
015/0608 v1

GoGet
GoFmt
GoBuildNull
GoBuild
GoRelease

example:
subst s1 s2 path
subst /usr/local/plan9port /usr/local/plan9 /usr/local/plan9/
*/

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

func subst(fp, s1, s2 string) error {
	var err error
	var fi os.FileInfo
	fi, err = os.Lstat(fp)
	if err != nil {
		return err
	}

	if fi.Mode().IsDir() {
		var ff []os.FileInfo
		ff, err = ioutil.ReadDir(fp)
		if err != nil {
			return err
		}
		var i os.FileInfo
		for _, i = range ff {
			err = subst(path.Join(fp, i.Name()), s1, s2)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// skip non-regular files
	if fi.Mode()&os.ModeType != 0 {
		return nil
	}

	var b1 []byte
	b1, err = ioutil.ReadFile(fp)
	if err != nil {
		return err
	}
	if !bytes.Contains(b1, []byte(s1)) {
		return nil
	}
	fmt.Fprintln(os.Stderr, fp)

	// BUG: file mode (permissions) lost
	var b2 []byte
	b2 = bytes.Replace(b1, []byte(s1), []byte(s2), -1)
	var tf *os.File
	tf, err = ioutil.TempFile("", "subst")
	if err != nil {
		return err
	}
	defer tf.Close()
	_, err = tf.Write(b2)
	if err != nil {
		return err
	}
	err = tf.Close()
	if err != nil {
		return err
	}
	err = os.Rename(tf.Name(), fp)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var err error
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "usage: subst string1 string2 path...\n")
		os.Exit(1)
	}

	for _, p := range os.Args[3:] {
		err = subst(path.Clean(p), os.Args[1], os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
