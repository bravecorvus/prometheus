package main

import (
    "fmt"
    "strconv"
    "strings"
    "time"
)

type yolo struct {
    one int
    two int
}

func errhandler(err error) {
    if err != nil {
        fmt.Println("You fucked up somewhere")
    }
}

func main() {
    currentyear, currentmonth, currentday := time.Now().Date()
    // thetime, err := time.Parse("2006:1:2:15:04", currentyear.String()+":"+currentmonth.String()+":"+currentday.String()+":"+"09:00")
    thestring := "09:00"
    split := strings.Split(thestring, "")
    fmt.Println(string(split[1]))
    hourldigit, err1 := strconv.Atoi(string(split[0]))
    errhandler(err1)
    hourrdigit, err2 := strconv.Atoi(string(split[1]))
    errhandler(err2)
    minuteldigit, err3 := strconv.Atoi(string(split[3]))
    errhandler(err3)
    minuterdigit, err4 := strconv.Atoi(string(split[4]))
    errhandler(err4)
    thetime := time.Date(currentyear, currentmonth, currentday, hourldigit+hourrdigit, minuteldigit+minuterdigit, 0, 0, time.UTC)
    // thetime.Year = currentyear
    // thetime.Month = currentmonth
    // thetime.Day = currentday
    // thetime.Add(currentday)
    // thetime.Add(currentmonth)
    // thetime.Add(currentyear)
    // if err != nil {
    //     fmt.Println("FUCKED UP")
    // }
    fmt.Println(thetime.Format("01/02 03:04:05PM '06 -0700"))
}
