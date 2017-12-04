package nixie

import (
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

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

func FindArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for what is mostly likely the Arduino device
	for _, f := range contents {
		if strings.Contains(f.Name(), "tty.usbserial") ||
			strings.Contains(f.Name(), "ttyUSB") {
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}
