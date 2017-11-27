// gpio is the speciic package that stores the functions called by Prometheus to run the bed vibrator (via go-rpio package)
package gpio

import (
	"fmt"
	"os"

	rpio "github.com/stianeikeland/go-rpio"
)

//Sends the signal to turn on the bed vibrator by sending a High (true) signal to GPIO 17
func VibOn() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	Input1 := rpio.Pin(5)
	Input1.Output()
	Input1.High()
	Input2 := rpio.Pin(6)
	Input2.Output()
	Input2.Low()
	Enable := rpio.Pin(17)
	Enable.Output()
	Enable.High()
}

//Sends the signal to turn off the bed vibrator by sending a Low (false) signal to GPIO 17
func VibOff() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	Input1 := rpio.Pin(5)
	Input1.Output()
	Input1.High()
	Input2 := rpio.Pin(6)
	Input2.Output()
	Input2.Low()
	Enable := rpio.Pin(17)
	Enable.Output()
	Enable.Low()
}
