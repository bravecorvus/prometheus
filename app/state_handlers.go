package app

import (
	"encoding/json"
	"net/http"

	"prometheus/store"
	"prometheus/utils"
)

// The frontend (public/index.html) fetches these URLs as if they were static
// files. After moving state into bbolt, we serve the same shapes from the
// in-memory app state so the client doesn't have to change.

func (app *App) AlarmsReadHandler(w http.ResponseWriter, r *http.Request) {
	// Legacy shape: array of {name, time, sound:"on"|"off", vibration:"on"|"off"}.
	type legacy struct {
		Name      string `json:"name"`
		Time      string `json:"time"`
		Sound     string `json:"sound"`
		Vibration string `json:"vibration"`
	}
	app.alarmsMu.Lock()
	out := make([]legacy, 0, NumAlarms)
	for _, a := range app.Alarms {
		out = append(out, legacy{
			Name:      a.Name,
			Time:      a.Alarmtime,
			Sound:     utils.ConvertBoolToOnOff(a.Sound),
			Vibration: utils.ConvertBoolToOnOff(a.Vibration),
		})
	}
	app.alarmsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (app *App) EmailReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(app.Email))
}

func (app *App) EnableEmailReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(boolText(app.EnableEmail)))
}

func (app *App) ColorsReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(app.Store.GetString(store.KeyColors)))
}

func (app *App) CustomSoundcardReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(boolText(app.CustomSoundCard)))
}

func (app *App) EnableLedReadHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(boolText(app.EnableLed)))
}

func boolText(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
