package nixie

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// CurrentTimeAsString returns the current time as a 6-character string (HHMMSS).
func CurrentTimeAsString() string {
	now := time.Now()
	return fmt.Sprintf("%02d%02d%02d", now.Hour(), now.Minute(), now.Second())
}

// FindArduino looks for /dev/ttyACM0 which is the Nixie clock Arduino.
func FindArduino() string {
	contents, _ := os.ReadDir("/dev")
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyACM0") {
			return "/dev/" + f.Name()
		}
	}
	return ""
}
