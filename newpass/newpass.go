/*

go mod init github.com/shoce/newpass
go mod tidy

GoFmt
GoBuildNull
GoBuild
GoRun

*/

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	CharSet      = "0123456789abcdefghjkmnpqrstvwxyz"
	FirstCharSet = "csvwxz"

	PasswordLength = 50
)

func main() {
	rand.Seed(time.Now().Unix())

	var pass strings.Builder

	random := rand.Intn(len(FirstCharSet))
	pass.WriteString(string(FirstCharSet[random]))

	for i := 0; i < PasswordLength-1; i++ {
		random := rand.Intn(len(CharSet))
		pass.WriteString(string(CharSet[random]))
	}

	fmt.Println(pass.String())
}
