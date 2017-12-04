package nixie

import (
	"strconv"
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
