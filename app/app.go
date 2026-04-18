package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"prometheus/nixie"
	"prometheus/structs"
	"prometheus/utils"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

const NumAlarms = 4

type App struct {
	Alarms         [NumAlarms]structs.Alarm
	EnableEmail    bool
	Email          string
	Soundname      string
	CustomSoundCard bool
	EnableLed      bool
	Red, Green, Blue string
	Options        serial.OpenOptions
	Port           io.ReadWriteCloser
	FoundNixie     bool
}

func (app *App) Initialize() {
	jsondata := structs.GetRawJson(utils.Pwd() + "/public/json/alarms.json")
	for i := 0; i < NumAlarms; i++ {
		app.Alarms[i].InitializeAlarms(jsondata, i)
	}

	var b bytes.Buffer
	if err := utils.Execute(&b, exec.Command("ls", utils.Pwd()+"/public/assets")); err != nil {
		fmt.Println(err)
	}

	app.Soundname = strings.TrimSpace(b.String())
	if err := os.WriteFile(utils.Pwd()+"/public/json/trackinfo", []byte(app.Soundname), 0644); err != nil {
		fmt.Println(err)
	}

	app.Email = utils.GetEmail()
	app.EnableEmail = utils.GetEnableEmail()
	app.CustomSoundCard = utils.UseCustomSoundCard()
	app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
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

// findAlarm returns a pointer to the alarm matching the given name, or nil.
func (app *App) findAlarm(name string) *structs.Alarm {
	for i := range app.Alarms {
		if app.Alarms[i].Name == name {
			return &app.Alarms[i]
		}
	}
	return nil
}

func (app *App) writeBackAlarms() {
	utils.WriteBackJson(app.Alarms[:], utils.Pwd()+"/public/json/alarms.json")
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
	if err := os.WriteFile(utils.Pwd()+"/public/json/trackinfo", []byte(app.Soundname), 0644); err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "File uploaded successfully: %s", header.Filename)
}

func (app *App) TimeHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	if alarm := app.findAlarm(r.FormValue("name")); alarm != nil {
		alarm.Alarmtime = r.FormValue("value")
		alarm.CurrentlyRunning = false
	}
	app.writeBackAlarms()
}

func (app *App) SoundHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	if alarm := app.findAlarm(r.FormValue("name")); alarm != nil {
		alarm.Sound = r.FormValue("value") == "on"
		alarm.CurrentlyRunning = false
	}
	app.writeBackAlarms()
}

func (app *App) VibrationHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	if alarm := app.findAlarm(r.FormValue("name")); alarm != nil {
		alarm.Vibration = r.FormValue("value") == "on"
		alarm.CurrentlyRunning = false
	}
	app.writeBackAlarms()
}

func (app *App) SnoozeHandler(w http.ResponseWriter, r *http.Request) {
	for i := range app.Alarms {
		if app.Alarms[i].CurrentlyRunning {
			app.Alarms[i].CurrentlyRunning = false
			app.Alarms[i].AddTime(app.Alarms[i].Alarmtime, "m", 10)
			app.writeBackAlarms()
			break
		}
	}
	http.Redirect(w, r, "/", 301)
}

func (app *App) EnableEmailHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.EnableEmail = r.FormValue("value") == "true"
	utils.WriteEnableEmail(r.FormValue("value"))
}

func (app *App) CustomSoundcardHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.CustomSoundCard = r.FormValue("value") == "true"
	utils.WriteCustomSoundCard(r.FormValue("value"))
}

func (app *App) NewEmailHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		return
	}
	utils.WriteEmail(r.FormValue("value"))
}

func (app *App) SubmitColorsHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.Red, app.Green, app.Blue = utils.ColorUpdate(r.FormValue("value"))
}

func (app *App) SubmitEnableLEDHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err.Error())
		return
	}
	app.EnableLed = r.FormValue("value") == "true"
	utils.WriteEnableLed(r.FormValue("value"))
}
