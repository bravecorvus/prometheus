package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"prometheus/config"
	"prometheus/store"
	"prometheus/structs"
	"prometheus/utils"
)

// stateDTO is the shape returned by GET /api/state and consumed by the React
// frontend on page load. It collects everything the UI needs in a single
// request so we don't fan out to half a dozen endpoints.
type stateDTO struct {
	Demo            bool             `json:"demo"`
	Alarms          []structs.Alarm  `json:"alarms"`
	Email           string           `json:"email"`
	EnableEmail     bool             `json:"enableEmail"`
	EnableLed       bool             `json:"enableLed"`
	CustomSoundCard bool             `json:"customSoundCard"`
	Colors          string           `json:"colors"`
	TrackInfo       string           `json:"trackInfo"`
}

func (app *App) StateHandler(w http.ResponseWriter, r *http.Request) {
	app.alarmsMu.Lock()
	alarms := make([]structs.Alarm, NumAlarms)
	copy(alarms, app.Alarms[:])
	app.alarmsMu.Unlock()

	resp := stateDTO{
		Demo:            config.DemoMode,
		Alarms:          alarms,
		Email:           app.Email,
		EnableEmail:     app.EnableEmail,
		EnableLed:       app.EnableLed,
		CustomSoundCard: app.CustomSoundCard,
		Colors:          app.Store.GetString(store.KeyColors),
		TrackInfo:       app.Soundname,
	}
	writeJSON(w, resp)
}

// alarmPatch is a partial update; only set fields are applied.
type alarmPatch struct {
	Time      *string `json:"time,omitempty"`
	Sound     *bool   `json:"sound,omitempty"`
	Vibration *bool   `json:"vibration,omitempty"`
}

func (app *App) AlarmPatchHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var patch alarmPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.alarmsMu.Lock()
	defer app.alarmsMu.Unlock()
	alarm := app.findAlarm(name)
	if alarm == nil {
		http.Error(w, "alarm not found", http.StatusNotFound)
		return
	}
	if patch.Time != nil {
		alarm.Alarmtime = *patch.Time
	}
	if patch.Sound != nil {
		alarm.Sound = *patch.Sound
	}
	if patch.Vibration != nil {
		alarm.Vibration = *patch.Vibration
	}
	alarm.CurrentlyRunning = false
	app.persistAlarms()
	writeJSON(w, alarm)
}

// SnoozeHandler shifts whichever alarm is currently ringing forward by 5
// minutes. It's the action the user triggers by tapping the Nixie display.
// The alarm's Sound/Vibration settings are preserved — only its time moves.
func (app *App) SnoozeHandler(w http.ResponseWriter, r *http.Request) {
	app.alarmsMu.Lock()
	defer app.alarmsMu.Unlock()
	for i := range app.Alarms {
		if app.Alarms[i].CurrentlyRunning {
			app.Alarms[i].CurrentlyRunning = false
			app.Alarms[i].AddTime(app.Alarms[i].Alarmtime, "m", 5)
			app.persistAlarms()
			break
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

type settingsPatch struct {
	Email           *string `json:"email,omitempty"`
	EnableEmail     *bool   `json:"enableEmail,omitempty"`
	EnableLed       *bool   `json:"enableLed,omitempty"`
	CustomSoundCard *bool   `json:"customSoundCard,omitempty"`
	Colors          *string `json:"colors,omitempty"`
}

func (app *App) SettingsPatchHandler(w http.ResponseWriter, r *http.Request) {
	var patch settingsPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if patch.Email != nil {
		app.Email = *patch.Email
		_ = app.Store.PutString(store.KeyEmail, app.Email)
	}
	if patch.EnableEmail != nil {
		app.EnableEmail = *patch.EnableEmail
		_ = app.Store.PutBool(store.KeyEnableEmail, app.EnableEmail)
	}
	if patch.EnableLed != nil {
		app.EnableLed = *patch.EnableLed
		_ = app.Store.PutBool(store.KeyEnableLed, app.EnableLed)
	}
	if patch.CustomSoundCard != nil {
		app.CustomSoundCard = *patch.CustomSoundCard
		_ = app.Store.PutBool(store.KeyCustomSoundcard, app.CustomSoundCard)
	}
	if patch.Colors != nil {
		app.Red, app.Green, app.Blue = utils.ParseHexToRGB(*patch.Colors)
		_ = app.Store.PutString(store.KeyColors, *patch.Colors)
	}
	w.WriteHeader(http.StatusNoContent)
}

// UploadHandler replaces the current audio file. The previous file is
// deleted to keep AssetsDir to one track at a time — the playback code only
// knows about a single Soundname.
func (app *App) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if app.Soundname != "" {
		_ = os.Remove(filepath.Join(app.AssetsDir, app.Soundname))
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Reject obvious path-escape attempts; the audio name ends up in shell
	// commands later, so be conservative.
	safe := filepath.Base(header.Filename)
	if safe == "" || safe == "." || safe == ".." || strings.ContainsAny(safe, `/\`) {
		http.Error(w, "invalid filename", http.StatusBadRequest)
		return
	}

	dst := filepath.Join(app.AssetsDir, safe)
	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, "unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Soundname = safe
	fmt.Fprintf(w, `{"trackInfo":%q}`, safe)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
