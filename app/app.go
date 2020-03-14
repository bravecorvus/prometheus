package app

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

type App struct {
	//Alarm object from the structs.Alarm{} struct
	Alarm1 structs.Alarm
	//Alarm object from the structs.Alarm{} struct
	Alarm2 structs.Alarm
	//Alarm object from the structs.Alarm{} struct
	Alarm3 structs.Alarm
	//Alarm object from the structs.Alarm{} struct
	Alarm4 structs.Alarm
	// Bool to keep track of whether the user wants to receive emails when the Pi's IP changes.
	// (If using DDNS, this is not necessary as the DDNS client will automatically update the relevant IP, and it will always be reachable with the same command)
	EnableEmail bool
	// The email address to send the IP change notification email to
	Email string
	// Alarm sound filename
	Soundname string
	// Used to see if user wants to use the cvlc command for use on a custom sound card, or the regular cvlc command
	CustomSoundCard bool
	// Used to see if user wants to enable LED's or not
	EnableLed bool
	// Used to tell what colors the user wants the LED to be on the clock
	Red, Green, Blue string
	Options          serial.OpenOptions
	Port             io.ReadWriteCloser
	FoundNixie       bool
}

//Initialization sequence
// 1. Grab the persistent alarm information from the alarms.json file
// 2. Use this data to store legitimate values of time, vibration, and sound into the 4 struct.Alarm objects
// 3. Grab the name of the current sound file via an "ls" shell command, and save it to the global variable Sound, and then write that information into ./public/json/trackinfo
// 4. Get the email of user to be used in the CheckIPChange() function
// 5. Get the persistent data of whether or not the user wants Prometheus to send emails when the IP changes. (Note, this is probably alot easier done through a dynamic DNS service which runs a background program to constanty check the IP of the Pi, updates a domain name server, and then you can access that as a link such as myclockname.ddns.net:3000
func (app *App) Initialize() {

	//Save the JSON alarms configurations into the mold
	jsondata := structs.GetRawJson(utils.Pwd() + "/public/json/alarms.json")
	app.Alarm1.InitializeAlarms(jsondata, 0)
	app.Alarm2.InitializeAlarms(jsondata, 1)
	app.Alarm3.InitializeAlarms(jsondata, 2)
	app.Alarm4.InitializeAlarms(jsondata, 3)

	var b bytes.Buffer
	if err := utils.Execute(&b,
		exec.Command("ls", utils.Pwd()+"/public/assets"),
	); err != nil {
		fmt.Println(err)
	}

	app.Soundname = strings.TrimSpace(b.String())
	d1 := []byte(app.Soundname)
	errrrrrrrrr := ioutil.WriteFile(utils.Pwd()+"/public/json/trackinfo", d1, 0644)
	if errrrrrrrrr != nil {
		fmt.Println(errrrrrrrrr)
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

	// Sleep since we need to wait for the Nixie clock to go online before starting to send it the time
	time.Sleep(20 * time.Second)

	// Open the serial USB port to communicate with the clock.
	// var Port io.ReadWriteCloser
	// if Options.PortName != "" {
	port, err := serial.Open(app.Options)
	if err != nil {
		app.FoundNixie = false
	} else {
		app.FoundNixie = true
		defer app.Port.Close()
	}
	app.Port = port

}

//Handle the AJAX file upload from the client. Deletes the old file, and saves the new name of the sound from the header.Filename
func (app *App) UploadHandler(w http.ResponseWriter, r *http.Request) {
	//Delete the old alarm sound via shell command process rm public/assets/sound_name.extension
	rmerror := exec.Command("rm", utils.Pwd()+"/public/assets/"+app.Soundname).Run()
	if rmerror != nil {
		fmt.Println(rmerror.Error())
	}
	file, header, err := r.FormFile("audio")
	//Set the Soundname attribute to the new soundname
	app.Soundname = header.Filename

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	//Write the soundname to ./public/assets/sound_name.extension
	out, err1 := os.Create(utils.Pwd() + "/public/assets/" + header.Filename)

	if err1 != nil {
		fmt.Fprintf(w, "Unable to upload the file")
	}
	defer out.Close()
	_, err2 := io.Copy(out, file)
	if err2 != nil {
		fmt.Fprintln(w, err)
	}
	fmt.Fprintf(w, "File uploaded successfully :")

	var b bytes.Buffer
	if err := utils.Execute(&b,
		exec.Command("ls", utils.Pwd()+"/public/assets"),
	); err != nil {
		fmt.Println(err)
	}

	// If a new file got uploaded, make sure this gets reflected in the program var Soundname and also write out the new name into ./public/json/trackinfo
	app.Soundname = strings.TrimSpace(b.String())
	d1 := []byte(app.Soundname)
	errrrrrrrrr := ioutil.WriteFile(utils.Pwd()+"/public/json/trackinfo", d1, 0644)
	if errrrrrrrrr != nil {
		fmt.Println(errrrrrrrrr)
	}

	fmt.Fprintf(w, header.Filename)
}

func (app *App) TimeHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
		os.Exit(1)
	}
	name := r.FormValue("name")
	time := r.FormValue("value")
	//Check to see which alarm the user is actually trying to modify, and modify the correct internally stored time
	if name == "alarm1" {
		app.Alarm1.Alarmtime = time
		app.Alarm1.CurrentlyRunning = false
	} else if name == "alarm2" {
		app.Alarm2.Alarmtime = time
		app.Alarm2.CurrentlyRunning = false
	} else if name == "alarm3" {
		app.Alarm3.Alarmtime = time
		app.Alarm3.CurrentlyRunning = false
	} else if name == "alarm4" {
		app.Alarm4.Alarmtime = time
		app.Alarm4.CurrentlyRunning = false
	}
	//Write back the alarm data back to the ./public/json/alarms.json so Prometheus could retreive the data after a restart
	utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
}

func (app *App) SoundHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
		os.Exit(1)
	}
	name := r.FormValue("name")
	sound := r.FormValue("value")
	var boolsound bool
	if sound == "on" {
		boolsound = true
	} else {
		boolsound = false
	}

	//Check to see which alarm the user is actually trying to modify, and modify the correct internally stored sound
	if name == "alarm1" {
		app.Alarm1.Sound = boolsound
		app.Alarm1.CurrentlyRunning = false
	} else if name == "alarm2" {
		app.Alarm2.Sound = boolsound
		app.Alarm2.CurrentlyRunning = false
	} else if name == "alarm3" {
		app.Alarm3.Sound = boolsound
		app.Alarm3.CurrentlyRunning = false
	} else if name == "alarm4" {
		app.Alarm4.Sound = boolsound
		app.Alarm4.CurrentlyRunning = false
	}
	//Write back the alarm data back to the ./public/json/alarms.json so Prometheus could retreive the data after a restart
	utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
}

func (app *App) VibrationHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
		os.Exit(1)
	}
	name := r.FormValue("name")
	vibration := r.FormValue("value")
	var boolvibration bool
	if vibration == "on" {
		boolvibration = true
	} else {
		boolvibration = false
	}
	//Check to see which alarm the user is actually trying to modify, and modify the correct internally stored vibration
	if name == "alarm1" {
		app.Alarm1.Vibration = boolvibration
		app.Alarm1.CurrentlyRunning = false
	} else if name == "alarm2" {
		app.Alarm2.Vibration = boolvibration
		app.Alarm2.CurrentlyRunning = false
	} else if name == "alarm3" {
		app.Alarm3.Vibration = boolvibration
		app.Alarm3.CurrentlyRunning = false
	} else if name == "alarm4" {
		app.Alarm4.Vibration = boolvibration
		app.Alarm4.CurrentlyRunning = false
	}
	//Write back the alarm data back to the ./public/json/alarms.json so Prometheus could retreive the data after a restart
	utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
}

func (app *App) SnoozeHandler(w http.ResponseWriter, r *http.Request) {
	//Using the AddTime() function, add 10 minutes to the currently running alarm, turn off the currently running alarm, and write back the correct configuration back to ./public/json/alarms.json
	if app.Alarm1.CurrentlyRunning {
		app.Alarm1.CurrentlyRunning = false
		app.Alarm1.AddTime(app.Alarm1.Alarmtime, "m", 10)
		utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
	} else if app.Alarm2.CurrentlyRunning {
		app.Alarm2.CurrentlyRunning = false
		app.Alarm2.AddTime(app.Alarm2.Alarmtime, "m", 10)
		utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
	} else if app.Alarm3.CurrentlyRunning {
		app.Alarm3.CurrentlyRunning = false
		app.Alarm3.AddTime(app.Alarm3.Alarmtime, "m", 10)
		utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
	} else if app.Alarm4.CurrentlyRunning {
		app.Alarm4.CurrentlyRunning = false
		app.Alarm4.AddTime(app.Alarm4.Alarmtime, "m", 10)
		utils.WriteBackJson(app.Alarm1, app.Alarm2, app.Alarm3, app.Alarm4, utils.Pwd()+"/public/json/alarms.json")
	}
	http.Redirect(w, r, "/", 301)
}

func (app *App) EnableEmailHandler(w http.ResponseWriter, r *http.Request) {

	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
		os.Exit(1)
	}
	value := r.FormValue("value")
	if value == "true" {
		app.EnableEmail = true
		utils.WriteEnableEmail("true")
	} else {
		app.EnableEmail = false
		utils.WriteEnableEmail("false")
	}
}

func (app *App) CustomSoundcardHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
		os.Exit(1)
	}
	value := r.FormValue("value")
	if value == "true" {
		app.EnableEmail = true
		utils.WriteCustomSoundCard("true")
	} else {
		app.EnableEmail = false
		utils.WriteCustomSoundCard("false")
	}
}

func (app *App) NewEmailHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr)
	}
	value := r.FormValue("value")
	utils.WriteEmail(value)
}

func (app *App) SubmitColorsHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
	}
	value := r.FormValue("value")
	app.Red, app.Green, app.Blue = utils.ColorUpdate(value)
}

func (app *App) SubmitEnableLEDHandler(w http.ResponseWriter, r *http.Request) {
	erawr := r.ParseForm()
	if erawr != nil {
		fmt.Println(erawr.Error())
	}
	value := r.FormValue("value")
	if value == "true" {
		app.EnableLed = true
		utils.WriteEnableLed("true")
	} else {
		app.EnableLed = false
		utils.WriteEnableLed("false")
	}
}
