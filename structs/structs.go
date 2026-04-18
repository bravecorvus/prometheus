package structs

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// JsonAlarms is the JSON representation of an alarm stored in alarms.json.
type JsonAlarms struct {
	JsonName      string `json:"name"`
	JsonTime      string `json:"time"`
	JsonSound     string `json:"sound"`
	JsonVibration string `json:"vibration"`
}

// Alarm is how Prometheus stores alarms internally.
type Alarm struct {
	Name             string
	Alarmtime        string
	Sound            bool
	Vibration        bool
	CurrentlyRunning bool
}

// GetRawJson reads and parses the alarms configuration from alarms.json.
func GetRawJson(filepath string) []JsonAlarms {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var alarms []JsonAlarms
	if err := json.Unmarshal(raw, &alarms); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return alarms
}

// InitializeAlarms populates an Alarm from the JSON data at the given index.
func (a *Alarm) InitializeAlarms(jsondata []JsonAlarms, index int) {
	a.Name = jsondata[index].JsonName
	a.Alarmtime = jsondata[index].JsonTime
	a.Sound = jsondata[index].JsonSound == "on"
	a.Vibration = jsondata[index].JsonVibration == "on"
	a.CurrentlyRunning = false
}

// AddTime adds the given amount of hours/minutes/seconds to the alarm time.
func (a *Alarm) AddTime(originaltime string, hms string, byhowmuch int) {
	currenttime, _ := time.Parse("15:04", originaltime)
	var updatedtime time.Time
	switch hms {
	case "h":
		updatedtime = currenttime.Add(time.Duration(byhowmuch) * time.Hour)
	case "m":
		updatedtime = currenttime.Add(time.Duration(byhowmuch) * time.Minute)
	case "s":
		updatedtime = currenttime.Add(time.Duration(byhowmuch) * time.Second)
	}
	a.Alarmtime = updatedtime.Format("15:04")
}
