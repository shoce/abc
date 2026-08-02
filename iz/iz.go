package main
import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)
const (
	NL = "\n"
)
var (
	F = fmt.Sprintf
	pout = fmt.Print
)
func perr(msg string) { fmt.Fprint(os.Stderr, msg+NL) }
func iz(root string) (total int64) {
	filepath.WalkDir(
		root, 
		func(p string, d fs.DirEntry, err error) error {
			if err != nil { perr(F("ERROR WalkDir [%s] %v", p, err)) ; return nil }
			info, err := d.Info()
			if err != nil { perr(F("ERROR WalkDir Info [%s] %v", p, err)) ; return nil }
			total += info.Size()
			return nil
		},
	)
	return total
}
func commas(n int64) string {
	s := strconv.FormatInt(n, 10)
	var neg bool
	if len(s)>0 && s[0]=='-' { neg = true }
	if neg { s = s[1:] }
	if len(s) <= 3 { if neg { s = "-" + s } ; return s }
	lead := len(s) % 3
	if lead == 0 { lead = 3 }
	out := s[:lead]
	for i:=lead ; i<len(s) ; i+=3 { out += "," + s[i:i+3] }
	if neg { out = "-" + out }
	return out
}
func main() {
	args := os.Args[1:]
	if len(args)==0 { args = []string{"."} }
	var total int64
	for _, a := range args {
		asz := iz(a)
		total += asz
		pout(F("<%s,> [%s]\n", commas(asz), a))
	}
	pout(F("<%s,> total\n", commas(total)))
}

