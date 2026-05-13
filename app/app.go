package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"prometheus/config"
	"prometheus/nixie"
	"prometheus/store"
	"prometheus/structs"
	"prometheus/utils"
	"strings"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

const NumAlarms = 4

type App struct {
	Alarms           [NumAlarms]structs.Alarm
	alarmsMu         sync.Mutex
	EnableEmail      bool
	Email            string
	Soundname        string
	CustomSoundCard  bool
	EnableLed        bool
	Red, Green, Blue string
	Options          serial.OpenOptions
	Port             io.ReadWriteCloser
	FoundNixie       bool
	Store            *store.Store
}

func (app *App) Initialize(s *store.Store) {
	app.Store = s

	app.loadAlarmsFromStore()

	var b bytes.Buffer
	if err := utils.Execute(&b, exec.Command("ls", utils.Pwd()+"/public/assets")); err != nil {
		fmt.Println(err)
	}

	app.Soundname = strings.TrimSpace(b.String())
	if err := os.WriteFile(utils.Pwd()+"/public/json/trackinfo", []byte(app.Soundname), 0o644); err != nil {
		fmt.Println(err)
	}

	app.Email = s.GetString(store.KeyEmail)
	app.EnableEmail = s.GetBool(store.KeyEnableEmail)
	app.CustomSoundCard = s.GetBool(store.KeyCustomSoundcard)
	app.Red, app.Green, app.Blue = utils.ParseHexToRGB(s.GetString(store.KeyColors))
	app.EnableLed = s.GetBool(store.KeyEnableLed)

	if config.DemoMode {
		app.FoundNixie = false
		return
	}

	app.Options.PortName = nixie.FindArduino()
	app.Options.BaudRate = 115200
	app.Options.DataBits = 8
	app.Options.StopBits = 1
	app.Options.MinimumReadSize = 4

	time.Sleep(20 * time.Second)

	port, err := serial.Open(app.Options)
	if err != nil {
		app.FoundNixie = false
	} else {
		app.FoundNixie = true
		app.Port = port
	}
}

// loadAlarmsFromStore reads all alarms from bbolt into the fixed-size array.
// bbolt's ForEach returns keys in lexicographical order, which for the
// "alarm1".."alarm4" naming scheme gives natural ordering.
func (app *App) loadAlarmsFromStore() {
	loaded, err := app.Store.LoadAlarms()
	if err != nil {
		fmt.Println("load alarms:", err)
	}
	for i := 0; i < NumAlarms && i < len(loaded); i++ {
		app.Alarms[i] = loaded[i]
		app.Alarms[i].CurrentlyRunning = false
	}
}

// findAlarm returns a pointer to the alarm matching the given name, or nil.
// Caller must hold alarmsMu.
func (app *App) findAlarm(name string) *structs.Alarm {
	for i := range app.Alarms {
		if app.Alarms[i].Name == name {
			return &app.Alarms[i]
		}
	}
	return nil
}

func (app *App) persistAlarms() {
	if err := app.Store.SaveAlarms(app.Alarms[:]); err != nil {
		fmt.Println("persist alarms:", err)
	}
}

func (app *App) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := exec.Command("rm", utils.Pwd()+"/public/assets/"+app.Soundname).Run(); err != nil {
		fmt.Println(err.Error())
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	app.Soundname = header.Filename

	out, err := os.Create(utils.Pwd() + "/public/assets/" + header.Filename)
	if err != nil {
		fmt.Fprintf(w, "Unable to upload the file")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		fmt.Fprintln(w, err)
		return
	}

	var b bytes.Buffer
	if err := utils.Execute(&b, exec.Command("ls", utils.Pwd()+"/public/assets")); err != nil {
		fmt.Println(err)
	}

	app.Soundname = strings.TrimSpace(b.String())
	if err := os.WriteFile(utils.Pwd()+"/public/json/trackinfo", []byte(app.Soundname), 0o644); err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", header.Filename)
}

func (app *App) TimeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.alarmsMu.Lock()
	if alarm := app.findAlarm(r.FormValue("name")); alarm != nil {
		alarm.Alarmtime = r.FormValue("value")
		alarm.CurrentlyRunning = false
	}
	app.persistAlarms()
	app.alarmsMu.Unlock()
}

func (app *App) SoundHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.alarmsMu.Lock()
	if alarm := app.findAlarm(r.FormValue("name")); alarm != nil {
		alarm.Sound = r.FormValue("value") == "on"
		alarm.CurrentlyRunning = false
	}
	app.persistAlarms()
	app.alarmsMu.Unlock()
}

func (app *App) VibrationHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.alarmsMu.Lock()
	if alarm := app.findAlarm(r.FormValue("name")); alarm != nil {
		alarm.Vibration = r.FormValue("value") == "on"
		alarm.CurrentlyRunning = false
	}
	app.persistAlarms()
	app.alarmsMu.Unlock()
}

func (app *App) SnoozeHandler(w http.ResponseWriter, r *http.Request) {
	app.alarmsMu.Lock()
	for i := range app.Alarms {
		if app.Alarms[i].CurrentlyRunning {
			app.Alarms[i].CurrentlyRunning = false
			app.Alarms[i].AddTime(app.Alarms[i].Alarmtime, "m", 10)
			app.persistAlarms()
			break
		}
	}
	app.alarmsMu.Unlock()
	http.Redirect(w, r, "/", 301)
}

func (app *App) EnableEmailHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.EnableEmail = r.FormValue("value") == "true"
	if err := app.Store.PutBool(store.KeyEnableEmail, app.EnableEmail); err != nil {
		fmt.Println("save enable_email:", err)
	}
}

func (app *App) CustomSoundcardHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.CustomSoundCard = r.FormValue("value") == "true"
	if err := app.Store.PutBool(store.KeyCustomSoundcard, app.CustomSoundCard); err != nil {
		fmt.Println("save custom_soundcard:", err)
	}
}

func (app *App) NewEmailHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}
	app.Email = r.FormValue("value")
	if err := app.Store.PutString(store.KeyEmail, app.Email); err != nil {
		fmt.Println("save email:", err)
	}
}

func (app *App) SubmitColorsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	hexVal := r.FormValue("value")
	app.Red, app.Green, app.Blue = utils.ParseHexToRGB(hexVal)
	if err := app.Store.PutString(store.KeyColors, hexVal); err != nil {
		fmt.Println("save colors:", err)
	}
}

func (app *App) SubmitEnableLEDHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.EnableLed = r.FormValue("value") == "true"
	if err := app.Store.PutBool(store.KeyEnableLed, app.EnableLed); err != nil {
		fmt.Println("save enable_led:", err)
	}
}
