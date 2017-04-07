package main

import (
    // "fmt"
    // "strconv"
    // "strings"
    "github.com/stianeikeland/go-rpio"
    "time"
)

func main() {
    var Enable rpio.Pin
    var Input1 rpio.Pin
    var Input2 rpio.Pin
    Enable = rpio.Pin(17)
    Enable.Output()
    Input1 = rpio.Pin(5)
    Input1.Output()
    Input1.High()
    Input2 = rpio.Pin(6)
    Input2.Output()
    Input2.Low()
    Enable.High()
    time.Sleep(5 * time.Second)
    Enable.High()
}
