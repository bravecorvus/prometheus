//main.go

package main

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/stianeikeland/go-rpio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Enable rpio.Pin
var Input1 rpio.Pin
var Input2 rpio.Pin

type JsonAlarm struct {
	Name      string `json:"name"`
	Alarm     string `json:"time"`
	Sound     string `json:"sound"`
	Vibration string `json:"vibration"`
}

type Alarm struct {
	Name             string
	Alarmtime        string
	Sound            bool
	Vibration        bool
	CurrentlyRunning bool
}

var Alarm1 = Alarm
var Alarm2 = Alarm
var Alarm3 = Alarm
var Alarm4 = Alarm

func (argumentalarm *Alarm) initializeAlarms(filepath string) {
	jsonalarm, err1 := ioutil.ReadFile(filepath)
	if err1 != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	var alarm []JsonAlarm
	err2 := json.Unmarshal(jsonalarm, &alarm)
	if err2 != nil {
		fmt.Println("ERROR JSON")
		os.Exit(1)
	}
	argumentalarm.Name = alarm[0].Name
	argumentalarm.Alarmtime = alarm[0].Alarm
	if alarm[0].Sound == "on" {
		argumentalarm.Sound = true
	} else {
		argumentalarm.Sound = false
	}
	if alarm[0].Vibration == "on" {
		argumentalarm.Vibration = true
	} else {
		argumentalarm.Vibration = false
	}
	argumentalarm.CurrentlyRunning = false

}

func Errhandler(err error) {
	if err != nil {
		fmt.Println("You fucked up somewhere")
	}
}
func StringTimeToReadTime(arg string) time.Time {
	//To prevent getting wierd time discrepancies with the time (since I am only saving the time itself, I need to intialize using the current Date as well, or else it will be 00/00 TI:ME '00 )
	currentyear, currentmonth, currentday := time.Now().Date()
	thestring := arg
	split := strings.Split(thestring, "")
	fmt.Println(string(split[1]))
	hourldigit, err1 := strconv.Atoi(string(split[0]))
	Errhandler(err1)
	hourrdigit, err2 := strconv.Atoi(string(split[1]))
	Errhandler(err2)
	minuteldigit, err3 := strconv.Atoi(string(split[3]))
	Errhandler(err3)
	minuterdigit, err4 := strconv.Atoi(string(split[4]))
	Errhandler(err4)
	return time.Date(currentyear, currentmonth, currentday, hourldigit+hourrdigit, minuteldigit+minuterdigit, 0, 0, time.UTC)
}

func (arg *Alarm) addTime(originaltime string, hms string, byhowmuch int) { //takes originaltime, and adds byhowmuch hours/minutes/seconds, then returns the string
	thetime, err := time.Parse("15:04", originaltime)
	fmt.Println("addTime")
	Errhandler(err)
	var updatedtime = time.Now()
	switch {
	case hms == "h":
		updatedtime = thetime.Add(time.Duration(byhowmuch) * time.Hour)
	case hms == "m":
		updatedtime = thetime.Add(time.Duration(byhowmuch) * time.Minute)
	case hms == "s":
		updatedtime = thetime.Add(time.Duration(byhowmuch) * time.Second)
	}
	arg.Alarmtime = updatedtime.Format("15:04")
}

func OverTenMinutes(alarm string, current string) bool {
	fmt.Println("OverTenMinutes")
	timealarm := StringTimeToReadTime(alarm)
	timecurrent := time.Now()
	diff := timecurrent.Sub(timealarm)
	fmt.Println(diff.Minutes())
	if diff.Minutes() > 10 {
		return false
	} else {
		return true
	}
}

func Runsnooze(channel chan bool, readyforreload chan bool) {
	fmt.Println("Runsnooze")
	go http.HandleFunc("/snooze", func(w http.ResponseWriter, r *http.Request) {
		channel <- true
		<-readyforreload
		http.Redirect(w, r, "/", 301)
	})
}

func (alarm *Alarm) RunAlarm(currenttime string, wg *sync.WaitGroup) {
	fmt.Println("RunAlarm")
	defer wg.Done()
	if (alarm.Sound == false) && (alarm.Vibration == false) && (alarm.CurrentlyRunning == false) {
		return
	}
	vibcounter := 0 //used to count 0-20 during which vibrations are on and 21-40 which means vibrations are off

	alarm.CurrentlyRunning = true //Set the state of the alarm to true

	var itsbeentenminutes bool //Used to see if an alarm has been running for ten minutes. If So, turn off the alarm, and add 1 hour to the clock

	startedWithMusic := false //Basically, in the event that the music was turned off through a separate process, this will ensure it won't mess anything up
	cmd := exec.Command("cvlc", "./public/assets/alarm.m4a")
	if alarm.Sound == true {
		startedWithMusic = true
		cmd.Start()
	}
	snoozed := make(chan bool)
	readyforreload := make(chan bool)
	go Runsnooze(snoozed, readyforreload)
	for {
		itsbeentenminutes = OverTenMinutes(alarm.Alarmtime, time.Now().Format("15:04"))
		switch { //Special cases using gochannels to listen to special activities
		case <-snoozed: //Just got snoozed
			fmt.Println("snooze just got pressed")
			if startedWithMusic {
				cmd.Process.Kill()
			}
			alarm.CurrentlyRunning = false
			alarm.addTime(alarm.Alarmtime, "m", 10)
			var writeback sync.WaitGroup
			writeback.Add(1)
			path := "./public/json/" + alarm.Name
			fmt.Println("path is " + path)
			switch {
			case alarm.Name == "alarm1":
				writeBackJson(*alarm, Alarm2, Alarm3, Alarm4, path, &writeback)
			case alarm.Name == "alarm2":
				writeBackJson(Alarm1, *alarm, Alarm3, Alarm4, path, &writeback)
			case alarm.Name == "alarm3":
				writeBackJson(Alarm1, Alarm2, *alarm, Alarm4, path, &writeback)
			case alarm.Name == "alarm4":
				writeBackJson(Alarm1, Alarm2, Alarm3, *alarm, path, &writeback)
			}
			writeback.Wait()
			readyforreload <- true
			return
		case itsbeentenminutes == true:
			fmt.Println("itsbeentenminutes")
			alarm.CurrentlyRunning = false
			if startedWithMusic == true {
				cmd.Process.Kill()
			}
			alarm.addTime(alarm.Alarmtime, "h", 1)
			var writeback sync.WaitGroup
			writeback.Add(1)
			path := "./public/json/" + alarm.Name
			writeBackJson(*alarm, path, &writeback)
			writeback.Wait()
			return
		default:
			switch {
			case ((alarm.Sound == false) && (alarm.Vibration == false) && (alarm.CurrentlyRunning == true) && (startedWithMusic)):
				fmt.Println("sound false | vibration false | currentlyrunning true")
				alarm.CurrentlyRunning = false
				cmd.Process.Kill()
				return
			case ((alarm.Sound == false) && (alarm.Vibration == true) && (alarm.CurrentlyRunning == true)):
				fmt.Println("sound false | vibration true | currentlyrunning true")
				fmt.Println("vib1")
				if vibcounter == 0 {
					VibOn()
				} else if vibcounter == 200 {
					VibOff()
				} else if vibcounter == 400 {
					vibcounter = 0
				}
				vibcounter++
			// case ((alarm.Sound == true) && (alarm.Vibration == false) && (alarm.CurrentlyRunning == true)):
			// 	time.Sleep(5 * time.Nanosecond)
			case ((alarm.Sound == true) && (alarm.Vibration == true) && (alarm.CurrentlyRunning == true)):
				fmt.Println("sound true | vibration true | currentlyrunning true")
				fmt.Println("vib2")
				if vibcounter == 0 {
					VibOn()
				} else if vibcounter == 200 {
					VibOff()
				} else if vibcounter == 400 {
					vibcounter = 0
				}
				vibcounter++
			}
		}
	}
}

func convertBooltoString(arg bool) string {
	if arg == true {
		return "on"
	} else {
		return "off"
	}
}

func writeBackJson(Alarm1 Alarm, Alarm2 Alarm, Alarm3 Alarm, Alarm4 Alarm, filepath string, wg *sync.WaitGroup) {
	defer wg.Done()
	// fmt.Println("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"}\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"}\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"}\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	content := []byte("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"}\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"}\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"}\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	err := ioutil.WriteFile(filepath, content, 0644)
	if err != nil {
		fmt.Println("Error writing back JSON alarm file for " + filepath)
		os.Exit(1)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()
	out, err := os.Create("./public/assets/alarm.m4a")
	if err != nil {
		fmt.Fprintf(w, "Unable to upload the file")
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}
	fmt.Fprintf(w, "File uploaded successfully :")
	fmt.Fprintf(w, header.Filename)
	http.Redirect(w, r, "/", 301)
}

func init() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	Input1 = rpio.Pin(5)
	Input1.Output()
	Input1.High()
	Input2 = rpio.Pin(6)
	Input2.Output()
	Input2.Low()
	Enable = rpio.Pin(17)
	Enable.Output()
	Enable.Low()
}

func main() {
	// Initialize all 4 instances of alarm clocks
	Alarm1.initializeAlarms("./public/json/Alarm1.json")
	Alarm2.initializeAlarms("./public/json/Alarm2.json")
	Alarm3.initializeAlarms("./public/json/Alarm3.json")
	Alarm4.initializeAlarms("./public/json/Alarm4.json")

	// Create function that updates clock once a minute (used to see if any times match up)
	t := time.Now()
	currenttime := t.Format("15:04")
	// fmt.Println(currenttime)
	c := cron.New()
	c.AddFunc("@every 1m", func() {
		t = time.Now()
		currenttime = t.Format("15:04")
		bundledAlarms := [4]Alarm{Alarm1, Alarm2, Alarm3, Alarm4}
		for _, alarm := range bundledAlarms {
			// alarm.Alarmtime = currenttime
			if alarm.Alarmtime == currenttime {
				var runningalarm sync.WaitGroup
				runningalarm.Add(1)
				alarm.RunAlarm(currenttime, &runningalarm)
				runningalarm.Wait()
				break
				// now := time.Now()
				// now.Add(10 * time.Minute)
				// nowstring := now.Format("15:04")
				// alarm.Alarmtime = nowstring
			}
		}
	})
	c.Start()

	// Server index.html under //public/index.html
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/Alarm1time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm1.Alarmtime = r.FormValue("mytime1")
		var time1 sync.WaitGroup
		time1.Add(1)
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm1.json", &time1)
		time1.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm1sound", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("sound1")
		// fmt.Println(stringedinput)
		if len(stringedinput) == 0 {
			if Alarm1.CurrentlyRunning == true {
				Alarm1.CurrentlyRunning = false
				Alarm1.Sound = false
			} else if Alarm1.CurrentlyRunning == false {
				Alarm1.Sound = false
			}
		} else {
			Alarm1.Sound = true
		}
		var sound1 sync.WaitGroup
		sound1.Add(1)
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm1.json", &sound1)
		sound1.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm1vibration", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("vibration1")
		if len(stringedinput) == 0 {
			if Alarm1.CurrentlyRunning == true {
				Alarm1.CurrentlyRunning = false
				Alarm1.Vibration = false
			} else if Alarm1.CurrentlyRunning == false {
				Alarm1.Vibration = false
			}
		} else {
			Alarm1.Vibration = true
		}
		var vibration1 sync.WaitGroup
		vibration1.Add(1)
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm1.json", &vibration1)
		vibration1.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm2time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm2.Alarmtime = r.FormValue("mytime2")
		// fmt.Println(stringedinput)
		var time2 sync.WaitGroup
		time2.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm2.json", &time2)
		time2.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm2sound", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("sound2")
		if len(stringedinput) == 0 {
			if Alarm2.CurrentlyRunning == true {
				Alarm2.CurrentlyRunning = false
				Alarm2.Sound = false
			} else if Alarm2.CurrentlyRunning == false {
				Alarm2.Sound = false
			}
		} else {
			Alarm2.Sound = true
		}
		var sound2 sync.WaitGroup
		sound2.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm2.json", &sound2)
		sound2.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm2vibration", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("vibration2")
		if len(stringedinput) == 0 {
			if Alarm2.CurrentlyRunning == true {
				Alarm2.CurrentlyRunning = false
				Alarm2.Vibration = false
			} else if Alarm2.CurrentlyRunning == false {
				Alarm2.Vibration = false
			}
		} else {
			Alarm2.Vibration = true
		}
		var vibration2 sync.WaitGroup
		vibration2.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm2.json", &vibration2)
		vibration2.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm3time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm3.Alarmtime = r.FormValue("mytime3")
		var time3 sync.WaitGroup
		time3.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm3.json", &time3)
		time3.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm3sound", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("sound3")
		if len(stringedinput) == 0 {
			if Alarm3.CurrentlyRunning == true {
				Alarm3.CurrentlyRunning = false
				Alarm3.Sound = false
			} else if Alarm3.CurrentlyRunning == false {
				Alarm3.Sound = false
			}
		} else {
			Alarm3.Sound = true
		}
		var sound3 sync.WaitGroup
		sound3.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm3.json", &sound3)
		sound3.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm3vibration", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("vibration3")
		if len(stringedinput) == 0 {
			if Alarm3.CurrentlyRunning == true {
				Alarm3.CurrentlyRunning = false
				Alarm3.Vibration = false
			} else if Alarm3.CurrentlyRunning == false {
				Alarm3.Vibration = false
			}
		} else {
			Alarm3.Vibration = true
		}
		var vibration3 sync.WaitGroup
		vibration3.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm3.json", &vibration3)
		vibration3.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm4time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm4.Alarmtime = r.FormValue("mytime4")
		var time4 sync.WaitGroup
		time4.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm4.json", &time4)
		time4.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm4sound", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("sound4")
		if len(stringedinput) == 0 {
			if Alarm4.CurrentlyRunning == true {
				Alarm4.CurrentlyRunning = false
				Alarm4.Sound = false
			} else if Alarm4.CurrentlyRunning == false {
				Alarm4.Sound = false
			}
		} else {
			Alarm4.Sound = true
		}
		var sound4 sync.WaitGroup
		sound4.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm4.json", &sound4)
		sound4.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/Alarm4vibration", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		stringedinput := r.FormValue("vibration4")
		if len(stringedinput) == 0 {
			if Alarm4.CurrentlyRunning == true {
				Alarm4.CurrentlyRunning = false
				Alarm4.Vibration = false
			} else if Alarm4.CurrentlyRunning == false {
				Alarm4.Vibration = false
			}
		} else {
			Alarm4.Vibration = true
		}
		var vibration4 sync.WaitGroup
		vibration4.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/Alarm4.json", &vibration4)
		vibration4.Wait()
		http.Redirect(w, r, "/", 401)
	})

	http.HandleFunc("/upload", uploadHandler)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func VibOn() {
	Enable.High()
}

func VibOff() {
	Enable.Low()
}
