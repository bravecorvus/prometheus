// prometheus is the main program logic that runs both the alarm logic and the web server hosting the front-end user interface for users to control the Prometheus. The main function is split into two parts: First is the part that runs two cron jobs: the first runs once every second to try to send the time + LED string to the nixie clock and the second once every minute to check if the current time matches a user supplied alarm (also running the releavnt alarm actions vibrate and or output sound based on the user set parameters). Then the second half deals with providing the web server functionality. The fileserver serves a plain HTML file whose client side scripting uses Vue.js, Bootstrap, and jQuery to read from the JSON files stored at ./public/json/...json to display the information. In this way, the program avoids having to implement a database. The rest of the code provides HTTP POST endpoints when the user submits information. For most of these, the logic involves unmarshaling the HTTP headers, extracting the information, updating the program variables, and writing back the changes to the JSON files (so that the changes survive a program crash, and when the user reloads the UI, it will read from the updated settings.)

package main

import (
	"fmt"
	"net/http"

	"prometheus/utils"

	"prometheus/app"

	"github.com/robfig/cron"
)

var (
	globals app.App
)

func init() {
	globals.Initialize()

}

// Main function
// Runs the cron job (checking once a minute at exactly the point when second is 00) to check if the current time matches the user supplied alarm time configuration, and then runs the alarm if an enabled alarm matches the time
// Runs a separate cron job (once every second) to send the current time as a string to the Nixie Clock through serial USB
// Also, main contains all the http HandleFunc's to deal with GET '/', POST '/time', POST '/sound', POST '/vibration', POST '/snooze', POST '/enableemail', POST '/newemail'
func main() {

	// Make sure to close it later.

	// Initialize all 4 instances of alarm clocks
	// Create function that updates clock once a minute (used to see if any times match up)
	c := cron.New()
	// Send relevant time clock over serial USB
	c.AddFunc("@every 1s", func() { globals.SendTime() })
	//Run the following once a minute
	//Check all 4 alarms to see if the current time matches any configurations
	c.AddFunc("0 * * * * *", func() { globals.AlarmLoop() })

	c.Start()

	// Server index.html under ./public/index.html
	fs := http.FileServer(http.Dir(utils.Pwd() + "/public"))
	http.Handle("/", fs)

	//Handle the AJAX post call to submit a new time for a certain alarm
	http.HandleFunc("/time", globals.TimeHandler)
	http.HandleFunc("/sound", globals.TimeHandler)
	http.HandleFunc("/vibration", globals.VibrationHandler)
	http.HandleFunc("/snooze", globals.SnoozeHandler)
	http.HandleFunc("/enableemail", globals.EnableEmailHandler)
	http.HandleFunc("/customsoundcard", globals.CustomSoundcardHandler)
	http.HandleFunc("/newemail", globals.NewEmailHandler)
	http.HandleFunc("/submitcolors", globals.SubmitColorsHandler)
	http.HandleFunc("/submitenableled", globals.SubmitEnableLEDHandler)
	http.HandleFunc("/upload", globals.UploadHandler)

	fmt.Println(http.ListenAndServe(":3000", nil))

}
