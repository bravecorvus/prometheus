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

var Alarm1 = Alarm{}
var Alarm2 = Alarm{}
var Alarm3 = Alarm{}
var Alarm4 = Alarm{}

func getRawJson(filepath string) []JsonAlarm {
	raw, err1 := ioutil.ReadFile(filepath)
	if err1 != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	var alarm []JsonAlarm
	json.Unmarshal(raw, &alarm)
	return alarm
}

func (argumentalarm *Alarm) initializeAlarms(jsondata []JsonAlarm, index int) {
	argumentalarm.Name = string(jsondata[index].Name)
	argumentalarm.Alarmtime = string(jsondata[index].Alarm)
	if string(jsondata[index].Sound) == "on" {
		argumentalarm.Sound = true
	} else {
		argumentalarm.Sound = false
	}
	if string(jsondata[index].Vibration) == "on" {
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

func (arg *Alarm) addTime(originaltime string, hms string, byhowmuch int) { //takes originaltime, and adds byhowmuch hours/minutes/seconds, then returns the string
	currenttime, err := time.Parse("15:04", originaltime)
	fmt.Println("addTime")
	Errhandler(err)
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

func OverTenMinutes(alarmtime string) bool {
	fmt.Println("OverTenMinutes")
	timealarm, err := time.Parse("15:04", alarmtime)
	Errhandler(err)
	timecurrent := time.Now()
	difference := timealarm.Minute() - timecurrent.Minute()
	fmt.Println("difference in minutes " + string(difference))
	if difference >= 10 {
		return false
	} else {
		return true
	}
}

func Runsnooze(alarm *Alarm, channel chan bool) {
	fmt.Println("Runsnooze")
	go http.HandleFunc("/snooze", func(w http.ResponseWriter, r *http.Request) {
		channel <- true
		alarm.addTime(alarm.Alarmtime, "m", 10)
		var wg sync.WaitGroup
		wg.Add(1)
		path := "./public/json/alarms.json"
		switch {
		case alarm.Name == "alarm1":
			writeBackJson(*alarm, Alarm2, Alarm3, Alarm4, path, &wg)
		case alarm.Name == "alarm2":
			writeBackJson(Alarm1, *alarm, Alarm3, Alarm4, path, &wg)
		case alarm.Name == "alarm3":
			writeBackJson(Alarm1, Alarm2, *alarm, Alarm4, path, &wg)
		case alarm.Name == "alarm4":
			writeBackJson(Alarm1, Alarm2, Alarm3, *alarm, path, &wg)
		}
		wg.Wait()
		http.Redirect(w, r, "/", 301)
	})
}

func (alarm *Alarm) RunAlarm(currenttime string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("RunAlarm")
	if (alarm.Sound == false) && (alarm.Vibration == false) {
		return
	}
	counter := 0                  //used to count 0-20 during which vibrations are on and 21-40 which means vibrations are off
	alarm.CurrentlyRunning = true //Set the state of the alarm to true
	var itsbeentenminutes bool    //Used to see if an alarm has been running for ten minutes. If So, turn off the alarm, and add 1 hour to the clock
	cmd := exec.Command("cvlc", "./public/assets/alarm.m4a")
	snoozed := make(chan bool)
	go Runsnooze(*alarm, snoozed)
	switch {
	case (alarm.Sound == true) && (alarm.Vibration == false):
		cmd.Start()
		itsbeentenminutes = OverTenMinutes(alarm.Alarmtime)
		for {
			fmt.Println("Sound true Vibration false")
			switch {
			case <-snoozed: //Just got snoozed
				cmd.Process.Kill()
				return
			case itsbeentenminutes:
				cmd.Process.Kill()
				alarm.addTime(alarm.Alarmtime, "h", 1)
				path := "./public/json/alarms.json"
				var writeback sync.WaitGroup
				writeback.Add(1)
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
				return
			default:
				continue
			}
		}
	case (alarm.Sound == false) && (alarm.Vibration == true):
		itsbeentenminutes = OverTenMinutes(alarm.Alarmtime)
		for {
			fmt.Println("Sound false Vibration true")
			switch {
			case <-snoozed: //Just got snoozed
				VibOff()
				return
			case itsbeentenminutes:
				VibOff()
				alarm.addTime(alarm.Alarmtime, "h", 1)
				var writeback sync.WaitGroup
				writeback.Add(1)
				path := "./public/json/alarms.json"
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
				// readyforreload <- true
				return

			default:
				if counter == 0 {
					VibOn()
					counter++
				} else if counter == 20 {
					VibOff()
					counter++
				} else if counter == 40 {
					counter = 0
				} else {
					counter++
				}

			}
		}
	case (alarm.Sound == true) && (alarm.Vibration == true):
		cmd.Start()
		itsbeentenminutes = OverTenMinutes(alarm.Alarmtime)
		for {
			fmt.Println("Sound true Vibration true")
			switch {
			case <-snoozed: //Just got snoozed
				cmd.Process.Kill()
				VibOff()
				return
			case itsbeentenminutes:
				cmd.Process.Kill()
				VibOff()
				alarm.addTime(alarm.Alarmtime, "h", 1)
				var writeback sync.WaitGroup
				writeback.Add(1)
				path := "./public/json/alarms.json"
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
				// readyforreload <- true
				return
			default:
				if counter == 0 {
					counter++
					VibOn()
				} else if counter == 20 {
					counter++
					VibOff()
				} else if counter == 40 {
					counter = 0
				} else {
					counter++
				}
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
	content := []byte("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"},\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"},\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"},\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
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
	jsondata := getRawJson("./public/json/alarms.json")
	Alarm1.initializeAlarms(jsondata, 0)
	Alarm2.initializeAlarms(jsondata, 1)
	Alarm3.initializeAlarms(jsondata, 2)
	Alarm4.initializeAlarms(jsondata, 3)
}

func main() {
	// Initialize all 4 instances of alarm clocks
	// Create function that updates clock once a minute (used to see if any times match up)
	t := time.Now()
	currenttime := t.Format("15:04")
	// fmt.Println(currenttime)
	c := cron.New()
	c.AddFunc("0 * * * * *", func() {
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

	http.HandleFunc("/alarm1time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm1.Alarmtime = r.FormValue("mytime1")
		var time1 sync.WaitGroup
		time1.Add(1)
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &time1)
		time1.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm1sound", func(w http.ResponseWriter, r *http.Request) {
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
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &sound1)
		sound1.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm1vibration", func(w http.ResponseWriter, r *http.Request) {
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
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &vibration1)
		vibration1.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm2time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm2.Alarmtime = r.FormValue("mytime2")
		// fmt.Println(stringedinput)
		var time2 sync.WaitGroup
		time2.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &time2)
		time2.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm2sound", func(w http.ResponseWriter, r *http.Request) {
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
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &sound2)
		sound2.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm2vibration", func(w http.ResponseWriter, r *http.Request) {
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
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &vibration2)
		vibration2.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm3time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm3.Alarmtime = r.FormValue("mytime3")
		var time3 sync.WaitGroup
		time3.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &time3)
		time3.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm3sound", func(w http.ResponseWriter, r *http.Request) {
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
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &sound3)
		sound3.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm3vibration", func(w http.ResponseWriter, r *http.Request) {
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
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &vibration3)
		vibration3.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm4time", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		Alarm4.Alarmtime = r.FormValue("mytime4")
		var time4 sync.WaitGroup
		time4.Add(1)
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &time4)
		time4.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm4sound", func(w http.ResponseWriter, r *http.Request) {
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
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &sound4)
		sound4.Wait()
		http.Redirect(w, r, "/", 301)
	})
	http.HandleFunc("/alarm4vibration", func(w http.ResponseWriter, r *http.Request) {
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
		go writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json", &vibration4)
		vibration4.Wait()
		http.Redirect(w, r, "/", 301)
	})

	http.HandleFunc("/upload", uploadHandler)
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}

func VibOn() {
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
	Enable.High()
}

func VibOff() {
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
