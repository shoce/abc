/*
it = internet time in beats, there are one thousand beats between midnights

usage: it [options]
options:
	y = print year since last millenium
	d = print date or day number since last january first
	s = print seconds or subbeats (there are one hundred subbeats in one beat)
	m = print military time, with date in month and day format, if printing date
	n = not to print time
	f = yd = print year, date and time
	ff = yds = print year, date, time and seconds/subbeats
	v = print string of year and day suitable for version tagging like 20.289
	vv = print string of year, day and beats suitable for version tagging like 20.289.499
	vvv = print string of year, day and beats and subbeats suitable for version tagging like 20.289.499.38

history:
2020/8/31 copy from tik
20/289 (2020/10/15) add options for version printing and not printing beats but date only
20/209 add options f and ff
21/1215 -m option to show date with months and time in military 24 hour

GoFmt
GoBuildNull
GoBuild
GoRun
*/

package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	Beat time.Duration = time.Duration(24) * time.Hour / 1000
)

var (
	tzBiel   *time.Location
	tnow     time.Time
	td0, ty0 time.Time
)

func militarytimefmt(td time.Duration) string {
	return fmt.Sprintf("%02d%02d", int(td/time.Hour), int((td%time.Hour)/time.Minute))
}

func secondsfmt(td time.Duration) string {
	return fmt.Sprintf("%02d", int((td%time.Minute)/time.Second))
}

func beatsfmt(td time.Duration) string {
	return fmt.Sprintf("%03d", int(td/Beat))
}

func subbeatsfmt(td time.Duration) string {
	return fmt.Sprintf("%02d", int((td/(Beat/100))%100))
}

func dayfmt(td time.Duration) string {
	day := int(td/(time.Duration(24)*time.Hour)) + 1
	return fmt.Sprintf("%03d", day)
}

func monthdayfmt(td time.Duration) string {
	return ty0.Add(td).Format("0102")
}

func yearfmt(t time.Time, zeropad bool) (y string) {
	if zeropad {
		y = fmt.Sprintf("%03d", t.Year()%1000)
	} else {
		y = fmt.Sprintf("%d", t.Year()%1000)
	}
	return y
}

func init() {
	tzBiel = time.FixedZone("Biel", 60*60)
	tnow = time.Now().In(tzBiel)
	td0 = time.Date(tnow.Year(), tnow.Month(), tnow.Day(), 0, 0, 0, 0, tzBiel)
	ty0 = time.Date(tnow.Year(), 1, 1, 0, 0, 0, 0, tzBiel)
}

func main() {
	var ptime, pseconds, pmilitarytime bool
	var pdate, pmonthday bool
	var pyear bool
	var pversion, plongversion, plonglongversion bool
	ptime = true
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "y") {
			pyear = true
		}
		if strings.Contains(arg, "d") {
			pdate = true
		}
		if strings.Contains(arg, "m") {
			pmonthday = true
			pmilitarytime = true
		}
		if strings.Contains(arg, "s") {
			pseconds = true
		}
		if strings.Contains(arg, "n") {
			ptime = false
			pseconds = false
		}
		if strings.Contains(arg, "f") {
			pyear = true
			pdate = true
			ptime = true
		}
		if strings.Count(arg, "f") > 1 {
			pseconds = true
		}
		if strings.Contains(arg, "v") {
			pversion = true
		}
		if strings.Count(arg, "v") > 1 {
			plongversion = true
		}
		if strings.Count(arg, "v") > 2 {
			plonglongversion = true
		}
	}

	td := time.Since(td0)

	if pversion || plongversion {
		var version string
		if pmonthday {
			version = fmt.Sprintf("%s.%s", yearfmt(tnow, true), monthdayfmt(time.Since(ty0)))
		} else {
			version = fmt.Sprintf("%s.%s", yearfmt(tnow, true), dayfmt(time.Since(ty0)))
		}
		if plongversion {
			if pmilitarytime {
				version += "." + militarytimefmt(td)
			} else {
				version += "." + beatsfmt(td)
			}
		}
		if plonglongversion {
			if pmilitarytime {
				version += "." + secondsfmt(td)
			} else {
				version += "." + subbeatsfmt(td)
			}
		}
		fmt.Println(version)
		return
	}

	var s string
	if ptime {
		if pmilitarytime {
			s += "@" + militarytimefmt(td)
		} else {
			s += "@" + beatsfmt(td)
		}
	}
	if pseconds {
		if pmilitarytime {
			s += "." + secondsfmt(td)
		} else {
			s += "." + subbeatsfmt(td)
		}
	}
	if pdate {
		if pmonthday {
			s = monthdayfmt(time.Since(ty0)) + s
		} else {
			s = dayfmt(time.Since(ty0)) + s
		}
	}
	if pyear {
		s = yearfmt(tnow, false) + "/" + s
	}

	fmt.Println(s)
}
