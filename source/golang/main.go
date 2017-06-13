//main.go
//TODO: Add ajax function handlers for time, sound, and vibration

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
	"time"
)

type jsonAlarms struct {
	JsonName      string `json:"name"`
	JsonTime      string `json:"time"`
	JsonSound     string `json:"sound"`
	JsonVibration string `json:"vibration"`
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
var Soundname string
var Playsound = exec.Command("cvlc", "\"./public/assets/"+Soundname+"\"")

func VibOn() {
	fmt.Println("VibOn")
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	Input1 := rpio.Pin(5)
	Input1.Output()
	Input1.High()
	Input2 := rpio.Pin(6)
	Input2.Output()
	Input2.Low()
	Enable := rpio.Pin(17)
	Enable.Output()
	Enable.High()
}

func VibOff() {
	fmt.Println("VibOff")
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	Input1 := rpio.Pin(5)
	Input1.Output()
	Input1.High()
	Input2 := rpio.Pin(6)
	Input2.Output()
	Input2.Low()
	Enable := rpio.Pin(17)
	Enable.Output()
	Enable.Low()
}

func getRawJson(filepath string) []jsonAlarms {
	raw, err1 := ioutil.ReadFile(filepath)
	if err1 != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	var alarm []jsonAlarms
	err2 := json.Unmarshal(raw, &alarm)
	if err2 != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	return alarm
}

func (argumentalarm *Alarm) initializeAlarms(jsondata []jsonAlarms, index int) {
	argumentalarm.Name = string(jsondata[index].JsonName)
	argumentalarm.Alarmtime = string(jsondata[index].JsonTime)
	if string(jsondata[index].JsonSound) == "on" {
		argumentalarm.Sound = true
	} else {
		argumentalarm.Sound = false
	}
	if string(jsondata[index].JsonVibration) == "on" {
		argumentalarm.Vibration = true
	} else {
		argumentalarm.Vibration = false
	}
	argumentalarm.CurrentlyRunning = false

}

func Errhandler(err error) {
	if err != nil {
		fmt.Println("ERROR")
	}
}

func (arg *Alarm) addTime(originaltime string, hms string, byhowmuch int) { //takes originaltime, and adds byhowmuch hours/minutes/seconds, then returns the string
	currenttime, _ := time.Parse("15:04", originaltime)
	fmt.Println("before fixed snooze time", currenttime)
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
	fmt.Println("fixed snooze time", arg.Alarmtime)
}

func OverTenMinutes(alarmtime string) bool {
	// fmt.Println("OverTenMinutes")
	year, month, day := time.Now().Date()
	var hour int
	var minutes int
	if string([]rune(alarmtime)[0]) == "0" {
		hour, _ = strconv.Atoi(string([]rune(alarmtime)[1:2]))
	} else {
		hour, _ = strconv.Atoi(string([]rune(alarmtime)[0:2]))
	}

	if string([]rune(alarmtime)[3]) == "0" {
		minutes, _ = strconv.Atoi(string([]rune(alarmtime)[4]))
	} else {
		minutes, _ = strconv.Atoi(string([]rune(alarmtime)[3:]))
	}

	dadatetime := time.Date(int(year), month, int(day), hour, minutes, 0, 0, time.Local)
	fmt.Print("alarm time is ")
	fmt.Println(dadatetime)
	timecurrent := time.Now()
	fmt.Print("current time is ")
	fmt.Println(timecurrent)
	difference := time.Date(int(year), month, int(day), hour, minutes, 0, 0, time.Local).Minute() - timecurrent.Minute()
	fmt.Println("Difference is", difference)
	if difference == 10 {
		return true
	} else {
		return false
	}
}

func convertBooltoString(arg bool) string {
	if arg {
		return "on"
	} else {
		return "off"
	}
}

func writeBackJson(Alarm1 Alarm, Alarm2 Alarm, Alarm3 Alarm, Alarm4 Alarm, filepath string) {
	// fmt.Println("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"}\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"}\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"}\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	content := []byte("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"},\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"},\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"},\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	err := ioutil.WriteFile(filepath, content, 0644)
	if err != nil {
		fmt.Println("Error writing back JSON alarm file for " + filepath)
		os.Exit(1)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("audio")
	//_, filename, err := r.FormFile("filename")
	fmt.Println(header.Filename)
	//fmt.Println(header)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()
	var removefile []string
	removefile = append(removefile, "./public/assets"+Soundname)
	rm := exec.Command("rm", removefile...)
	errr := rm.Start()
	if errr != nil {
		fmt.Println("ERROR")
		os.Exit(1)
	}
	out, err1 := os.Create("./public/assets/" + header.Filename)
	Soundname = header.Filename

	if err1 != nil {
		fmt.Fprintf(w, "Unable to upload the file")
	}
	defer out.Close()
	_, err2 := io.Copy(out, file)
	if err2 != nil {
		fmt.Fprintln(w, err)
	}
	fmt.Fprintf(w, "File uploaded successfully :")
	fmt.Fprintf(w, header.Filename)
}

func init() {
	jsondata := getRawJson("./public/json/alarms.json")
	Alarm1.initializeAlarms(jsondata, 0)
	Alarm2.initializeAlarms(jsondata, 1)
	Alarm3.initializeAlarms(jsondata, 2)
	Alarm4.initializeAlarms(jsondata, 3)

	var assets []string
	assets = append(assets, "./public/assets/")
	ls := exec.Command("ls", assets...)
	cmdReader, err := ls.CombinedOutput()
	if err != nil {
		fmt.Println("ERROR")
	}
	Soundname = string(cmdReader[:])
}

func main() {
	// Initialize all 4 instances of alarm clocks
	// Create function that updates clock once a minute (used to see if any times match up)
	t := time.Now()
	currenttime := t.Format("15:04")
	c := cron.New()
	c.AddFunc("0 * * * * *", func() {
		breaktime := false
		duration := time.Second * 3
		t = time.Now()
		currenttime = t.Format("15:04")

		if Alarm1.Alarmtime == currenttime {
			fmt.Println("Alarm 1")
			Alarm1.CurrentlyRunning = true
			if Alarm1.Sound && Alarm1.Vibration {
				fmt.Println("Sound and Vibration")
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm1.CurrentlyRunning {
							fmt.Println("Breaking loop")
							breaktime = true
							break
						}
					}
					if breaktime {

						fmt.Println("breaking loop")
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						fmt.Println("Its been 10 minutes")
						Alarm1.CurrentlyRunning = false
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm1.Sound && !Alarm1.Vibration {
				fmt.Println("Sound")
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					if !Alarm1.CurrentlyRunning {
						fmt.Println("Breaking loop")
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						Alarm1.CurrentlyRunning = false
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm1.Sound && Alarm1.Vibration {
				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm1.CurrentlyRunning {
							breaktime = true
							fmt.Println("Breaking loop")
							break
						}
					}
					if breaktime {
						VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						fmt.Println("Its been ten minutes")
						Alarm1.CurrentlyRunning = false
						VibOff()
					} else {
						VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm1.CurrentlyRunning = false
			}

		} else if Alarm2.Alarmtime == currenttime {
			Alarm2.CurrentlyRunning = true
			if Alarm2.Sound && Alarm2.Vibration {
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm2.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm2.Sound && !Alarm2.Vibration {
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					if !Alarm2.CurrentlyRunning {
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm2.Sound && Alarm2.Vibration {
				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm2.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						VibOff()
					} else {
						VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm2.CurrentlyRunning = false
			}

		} else if Alarm3.Alarmtime == currenttime {
			Alarm3.CurrentlyRunning = true
			if Alarm3.Sound && Alarm3.Vibration {
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm3.Sound && !Alarm3.Vibration {
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					if !Alarm3.CurrentlyRunning {
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm3.Sound && Alarm3.Vibration {
				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						VibOff()
					} else {
						VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm3.CurrentlyRunning = false
			}

		} else if Alarm4.Alarmtime == currenttime {
			Alarm4.CurrentlyRunning = true
			if Alarm4.Sound && Alarm4.Vibration {
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						VibOff()
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm4.Sound && !Alarm4.Vibration {
				errrrror := Playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					if !Alarm4.CurrentlyRunning {
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						errrrrorkill := Playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm4.Sound && Alarm4.Vibration {
				for {
					VibOn()
					for i := 1; i <= 200; i++ {
						if !Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						VibOff()
					} else {
						VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm4.CurrentlyRunning = false
			}
		}
	})
	c.Start()

	// Server index.html under //public/index.html
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		name := r.FormValue("name")
		time := r.FormValue("value")
		//fmt.Println(name)
		if name == "alarm1" {
			Alarm1.Alarmtime = time
			Alarm1.CurrentlyRunning = false
		} else if name == "alarm2" {
			Alarm2.Alarmtime = time
			Alarm2.CurrentlyRunning = false
		} else if name == "alarm3" {
			Alarm3.Alarmtime = time
			Alarm3.CurrentlyRunning = false
		} else if name == "alarm4" {
			Alarm4.Alarmtime = time
			Alarm4.CurrentlyRunning = false
		}
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
	})

	http.HandleFunc("/sound", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		name := r.FormValue("name")
		sound := r.FormValue("value")
		//fmt.Println(name)
		var boolsound bool
		if sound == "on" {
			boolsound = true
		} else {
			boolsound = false
		}

		if name == "alarm1" {
			Alarm1.Sound = boolsound
			Alarm1.CurrentlyRunning = false
		} else if name == "alarm2" {
			Alarm2.Sound = boolsound
			Alarm2.CurrentlyRunning = false
		} else if name == "alarm3" {
			Alarm3.Sound = boolsound
			Alarm3.CurrentlyRunning = false
		} else if name == "alarm4" {
			Alarm4.Sound = boolsound
			Alarm4.CurrentlyRunning = false
		}
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
	})

	http.HandleFunc("/vibration", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		name := r.FormValue("name")
		vibration := r.FormValue("value")
		//fmt.Println(name)
		var boolvibration bool
		if vibration == "on" {
			boolvibration = true
		} else {
			boolvibration = false
		}
		if name == "alarm1" {
			Alarm1.Vibration = boolvibration
			Alarm1.CurrentlyRunning = false
		} else if name == "alarm2" {
			Alarm2.Vibration = boolvibration
			Alarm2.CurrentlyRunning = false
		} else if name == "alarm3" {
			Alarm3.Vibration = boolvibration
			Alarm3.CurrentlyRunning = false
		} else if name == "alarm4" {
			Alarm4.Vibration = boolvibration
			Alarm4.CurrentlyRunning = false
		}
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
	})

	http.HandleFunc("/snooze", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("snoozed")
		if Alarm1.CurrentlyRunning {
			Alarm1.CurrentlyRunning = false
			Alarm1.addTime(Alarm1.Alarmtime, "m", 10)
		} else if Alarm2.CurrentlyRunning {
			Alarm2.CurrentlyRunning = false
			Alarm2.addTime(Alarm2.Alarmtime, "m", 10)
		} else if Alarm3.CurrentlyRunning {
			Alarm3.CurrentlyRunning = false
			Alarm3.addTime(Alarm3.Alarmtime, "m", 10)
		} else if Alarm4.CurrentlyRunning {
			Alarm4.CurrentlyRunning = false
			Alarm4.addTime(Alarm4.Alarmtime, "m", 10)
		}
		http.Redirect(w, r, "/", 301)
	})

	http.HandleFunc("/upload", uploadHandler)
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
