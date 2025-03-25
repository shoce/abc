/*
GoFmt GoBuildNull GoBuild
*/

package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: htpasswd username password")
		os.Exit(1)
	}
	username, password := os.Args[1], os.Args[2]

	sum, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
	passwordHashed := string(sum)

	htpasswd := fmt.Sprintf("%s:%s", username, passwordHashed)
	fmt.Println(htpasswd)
	fmt.Println(base64.StdEncoding.EncodeToString([]byte(htpasswd)))
}
