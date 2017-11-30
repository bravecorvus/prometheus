// structs is the package that defines the various structs used by Prometheus, namely, JsonAlarms and Alarm as well as the the related functions that operate on the structs
package structs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

//JSON struct used when reading in the ./public/json/alarms.json to get the alarm configurations when main is first started
type JsonAlarms struct {
	JsonName      string `json:"name"`
	JsonTime      string `json:"time"`
	JsonSound     string `json:"sound"`
	JsonVibration string `json:"vibration"`
}

//How Prometheus stores the Alarms internally. all 4 alarms are of type "Alarm"
type Alarm struct {
	Name             string //Name of the alarm: one of the following: alarm1, alarm2, alarm3, or alarm4
	Alarmtime        string //Time of the alarm, stored as time; in 24 hour time (e.g. 11:00PM = 23:00).
	Sound            bool   //Sets whether Prometheus should play sound or not
	Vibration        bool   //Sets whether Prometheus should vibrate or not
	CurrentlyRunning bool   //Set to true when alarm is running, false when the user tells it to stop (i.e. Sound/Vibration toggle is turned off for current alarm, it has been more then 10 minutes, or the user pressed the snooze button)
}

//Get the alarms configurations from the ./public/json/alarms.json
func GetRawJson(filepath string) []JsonAlarms {
	raw, err1 := ioutil.ReadFile(filepath)
	if err1 != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	var alarm []JsonAlarms
	err2 := json.Unmarshal(raw, &alarm)
	if err2 != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	return alarm
}

//Populate the values of all 4 internally stored alarms using the values stored in the ./public/json/alarms.json
func (argumentalarm *Alarm) InitializeAlarms(jsondata []JsonAlarms, index int) {
	argumentalarm.Name = string(jsondata[index].JsonName)
	argumentalarm.Alarmtime = string(jsondata[index].JsonTime)
	if string(jsondata[index].JsonSound) == "on" {
		argumentalarm.Sound = true
	} else {
		argumentalarm.Sound = false
	}
	if string(jsondata[index].JsonVibration) == "on" {
		argumentalarm.Vibration = true
	} else {
		argumentalarm.Vibration = false
	}
	argumentalarm.CurrentlyRunning = false

}

//Function that adds 10 minutes to the currently running alarm. Necessary because internally, alarm times are stored as strings, rather than the time class
// Genrally called when you hit the snooze button on the front-end UI, which does a POST to '/snooze'
// takes originaltime, and adds byhowmuch hours/minutes/seconds, then returns the string
func (arg *Alarm) AddTime(originaltime string, hms string, byhowmuch int) {
	currenttime, _ := time.Parse("15:04", originaltime)
	var updatedtime = time.Now()
	switch {
	case hms == "h":
		updatedtime = currenttime.Add(time.Duration(byhowmuch) * time.Hour)
	case hms == "m":
		updatedtime = currenttime.Add(time.Duration(byhowmuch) * time.Minute)
	case hms == "s":
		updatedtime = currenttime.Add(time.Duration(byhowmuch) * time.Second)
	}
	arg.Alarmtime = updatedtime.Format("15:04")
}
