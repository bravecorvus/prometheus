package structs

import "time"

// Alarm is how Prometheus stores alarms internally and on disk.
// CurrentlyRunning is runtime-only and is excluded from JSON encoding so it
// never gets persisted to the store.
type Alarm struct {
	Name             string `json:"name"`
	Alarmtime        string `json:"time"`
	Sound            bool   `json:"sound"`
	Vibration        bool   `json:"vibration"`
	CurrentlyRunning bool   `json:"-"`
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
