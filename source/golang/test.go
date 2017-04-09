package main

import (
    "fmt"
    "time"
)

func main() {
    rn := time.Now()
    alarm, err := time.Parse("15:04", "11:11")
    if err != nil {
        fmt.Println("ERROR")
    }
    fmt.Println(alarm.Minute() - rn.Minute())
}
