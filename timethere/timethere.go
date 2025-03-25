// go run timethere.go

package main

import ("fmt"; "time")

func main() {
	const tfmt = "2006-01-02 15:04"
	var now, then time.Time
	var here, la *time.Location

	now = time.Now().Round(time.Second)
	here = now.Location()
	la, _ = time.LoadLocation("America/Los_Angeles")

	fmt.Println(now)
	fmt.Println(now.In(la))

	then, _ = time.ParseInLocation(tfmt, "2015-11-01 4:42", here)
	fmt.Println(then)
	fmt.Println(then.In(la))
}
