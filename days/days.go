/*
GoFmt
GoBuild
GoRun
*/

package main

import (
	"fmt"
	"strings"
	"time"
)

const (
	Cycle       = 28
	SEP         = " ; "
	DATESEP     = "/"
	TodayPrefix = "* "
	TodaySuffix = " *"
)

func dayfmt(td time.Duration) string {
	days := int(td / (time.Duration(24) * time.Hour))
	//return fmt.Sprintf("%d", days+1)
	return fmt.Sprintf("%d"+DATESEP+"%d", (days/Cycle)+1, (days%Cycle)+1)
}

func main() {
	tnow := time.Now()
	ty0 := time.Date(tnow.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	ty1 := time.Date(tnow.Year()+1, 1, 1, 0, 0, 0, 0, time.Local)

	for t := ty0; t.Before(ty1); t = t.Add(time.Duration(24) * time.Hour) {
		if t.Sub(ty0)%(Cycle*time.Duration(24)*time.Hour) == 0 {
			fmt.Println()
		}
		today := false
		if tnow.After(t) && tnow.Before(t.Add(time.Duration(24)*time.Hour)) {
			today = true
		}
		days := dayfmt(t.Sub(ty0))
		canondate := strings.ToLower(t.Format("Jan" + DATESEP + "2"))
		date := days + " " + canondate
		if today {
			date = TodayPrefix + date + TodaySuffix
		}
		fmt.Print(date + SEP)
	}
	fmt.Println()
}
