/*
GoFmt
GoRun
GoBuild
*/

package main

import (
	"fmt"
	"time"
)

const Cycle = 28

func dayfmt(td time.Duration) string {
	days := int(td / (time.Duration(24) * time.Hour))
	//return fmt.Sprintf("%d", days+1)
	return fmt.Sprintf("%d/%d", (days/Cycle)+1, (days%Cycle)+1)
}

func main() {
	tnow := time.Now()
	ty0 := time.Date(tnow.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	ty1 := time.Date(tnow.Year()+1, 1, 1, 0, 0, 0, 0, time.Local)

	var prefix string
	for t := ty0; t.Before(ty1); t = t.Add(time.Duration(24) * time.Hour) {
		if tnow.After(t) && tnow.Before(t.Add(time.Duration(24)*time.Hour)) {
			prefix = ">"
		} else {
			prefix = " "
		}
		fmt.Println(prefix, dayfmt(t.Sub(ty0)), t.Format("Jan/2"))
	}
}
