/*
GoGet
GoFmt
GoBuildNull
GoBuild
*/

package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const (
	NL = "\n"
)

var (
	F = fmt.Sprintf
	pout = fmt.Print
)

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		perr("USAGE htpasswd username password")
		os.Exit(1)
	}
	username, password := args[0], args[1]

	passsum, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		perr(F("ERROR %v", err))
		os.Exit(1)
	}
	passhash := string(passsum)

	htpasswd := F("%s:%s", username, passhash)
	pout(htpasswd+NL)
	pout(base64.StdEncoding.EncodeToString([]byte(htpasswd))+NL)
}

func perr(msg string) (int, error) {
	return fmt.Fprint(os.Stderr, msg+NL)
}

