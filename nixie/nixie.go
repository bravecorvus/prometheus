package nixie

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// Grabs the current time using the time library, and returns it as a string.
// Important that the string is of length 6 because the Arduino sketch uses the substring() method to extract parts of the entire string and uses char like indexing to grab the relevant information
// Hence, 1:2:3 will actually yield the string 010203
func CurrentTimeAsString() string {
	hour := time.Now().Hour()
	minute := time.Now().Minute()
	second := time.Now().Second()
	var stringhour, stringminute, stringsecond string
	ihour := strconv.Itoa(hour)
	iminute := strconv.Itoa(minute)
	isecond := strconv.Itoa(second)
	if hour < 10 {
		stringhour = "0" + ihour
	} else {
		stringhour = ihour
	}
	if minute < 10 {
		stringminute = "0" + iminute
	} else {
		stringminute = iminute
	}
	if second < 10 {
		stringsecond = "0" + isecond
	} else {
		stringsecond = isecond
	}

	return stringhour + stringminute + stringsecond
}

// Function to find the Arduino clock
// Specifically, on my Raspberry Pi, the Arduino device is named "/dev/ttyACM0" in the Pi
func FindArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for what is mostly likely the Arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyACM0") {
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}
