package main

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func main() {
	year, month, day := time.Now().Date()
	dastring := "02:01"
	var hour int
	var minutes int
	if string([]rune(dastring)[0]) == "0" {
		hour, _ = strconv.Atoi(string([]rune(dastring)[1:2]))
	} else {
		hour, _ = strconv.Atoi(string([]rune(dastring)[0:2]))
	}

	if string([]rune(dastring)[3]) == "0" {
		minutes, _ = strconv.Atoi(string([]rune(dastring)[4]))
	} else {
		minutes, _ = strconv.Atoi(string([]rune(dastring)[3:]))
	}

	fmt.Println("year: ", reflect.TypeOf(year))
	fmt.Println("month: ", reflect.TypeOf(int(month)))
	fmt.Println("day: ", reflect.TypeOf(day))
	fmt.Println("hour: ", hour)
	fmt.Println("minute: ", minutes)
	dadatetime := time.Date(int(year), month, int(day), hour, minutes, 0, 0, time.Local)
	fmt.Println(dadatetime)
}
