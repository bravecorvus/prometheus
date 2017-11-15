//main.go

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"./gpio"
	"./structs"
	"./utils"

	"github.com/gilgameshskytrooper/cron"
)

//Declare the alarms
var Alarm1 = structs.Alarm{}
var Alarm2 = structs.Alarm{}
var Alarm3 = structs.Alarm{}
var Alarm4 = structs.Alarm{}

//Declare the name of the alarm sound stored in ./public/assets/sound_name.extension
var Soundname string

//Declare the identity of the current wlan0 IP. Used to check if the IP changed.
var IP, NewIP string

//General error handler. I guess it wasn't used nearly as much as should to warrant it's existance, but its here nonetheless.
func Errhandler(err error) {
	if err != nil {
		fmt.Println("ERROR")
	}
}

//Function to check whether the alarm has been running for more than 10 minutes
//***NOT WORKING AS OF NOW, BUT DOES NOT NEGATIVELY AFFECT THE FUNCTIONALITY OF THE REST OF THE PROGRAM***
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

	//fmt.Print("alarm time is ")
	//fmt.Println(dadatetime)
	timecurrent := time.Now()
	//fmt.Print("current time is ")
	//fmt.Println(timecurrent)
	difference := time.Date(int(year), month, int(day), hour, minutes, 0, 0, time.Local).Minute() - timecurrent.Minute()
	//fmt.Println("Difference is", difference)
	if difference == 10 {
		return true
	} else {
		return false
	}
}

//Handle the AJAX file upload from the client. Deletes the old file, and saves the new name of the sound from the header.Filename
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	//Delete the old alarm sound via shell command process rm public/assets/sound_name.extension
	rmerror := exec.Command("rm", utils.Pwd()+"/public/assets/"+Soundname).Run()
	if rmerror != nil {
		fmt.Println(os.Stderr, rmerror)
	}
	file, header, err := r.FormFile("audio")
	//Set the Soundname attribute to the new soundname
	Soundname = header.Filename
	//_, filename, err := r.FormFile("filename")
	//fmt.Println(header.Filename)
	//fmt.Println(header)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	//Write the soundname to ./public/assets/sound_name.extension
	out, err1 := os.Create(utils.Pwd() + "/public/assets/" + header.Filename)

	//if err != nil {
	//fmt.Println("ERROR")
	//}

	if err1 != nil {
		fmt.Fprintf(w, "Unable to upload the file")
	}
	defer out.Close()
	_, err2 := io.Copy(out, file)
	if err2 != nil {
		fmt.Fprintln(w, err)
	}
	fmt.Fprintf(w, "File uploaded successfully :")
	//fmt.Fprintf(w, header.Filename)
}

//Initialization
func init() {
	// currentDir = utils.Pwd()
	//Save the JSON alarms configurations into the mold
	jsondata := structs.GetRawJson(utils.Pwd() + "/public/json/alarms.json")
	//Set up the intial values for the alarms using the values we get from above
	Alarm1.InitializeAlarms(jsondata, 0)
	Alarm2.InitializeAlarms(jsondata, 1)
	Alarm3.InitializeAlarms(jsondata, 2)
	Alarm4.InitializeAlarms(jsondata, 3)
	//Read the most recent IP address from ./public/json/ip
	IP = strings.TrimSpace(utils.GetIPFromFile())
	//Grab the name of the alarm sound file via ls command to the ./public/assets/ folder.
	var b bytes.Buffer
	if err := utils.Execute(&b,
		exec.Command("ls", utils.Pwd()+"/public/assets"),
	); err != nil {
		log.Fatalln(err)
	}
	//Since the ls stdout returns the string + a newline (\n), the following strips the newline from the string, and saves it as the Soundname variable
	Soundname = strings.TrimSpace(b.String())
	//fmt.Println(Soundname)
	d1 := []byte(Soundname)
	errrrrrrrrr := ioutil.WriteFile("initial", d1, 0644)
	if errrrrrrrrr != nil {
		fmt.Println(errrrrrrrrr)
	}
}

func main() {
	// Initialize all 4 instances of alarm clocks
	// Create function that updates clock once a minute (used to see if any times match up)
	t := time.Now()
	currenttime := t.Format("15:04")
	c := cron.New()
	//Run the following once a minute
	//Check all 4 alarms to see if the current time matches any configurations
	c.AddFunc("0 * * * * *", func() {
		breaktime := false
		duration := time.Second * 3
		t = time.Now()
		currenttime = t.Format("15:04")
		//fmt.Println("currenttime", currenttime)
		//fmt.Println("Alarm1Time", Alarm1.Alarmtime)

		if Alarm1.Alarmtime == currenttime {
			fmt.Println("YOLO1")
			NewIP = utils.GetIP()
			utils.Send(NewIP)
			//fmt.Println("Alarm 1")
			Alarm1.CurrentlyRunning = true

			if Alarm1.Sound && Alarm1.Vibration {
				//fmt.Println("Sound and Vibration")
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm1.CurrentlyRunning {
							//fmt.Println("Breaking loop")
							breaktime = true
							break
						}
					}
					if breaktime {

						//fmt.Println("breaking loop")
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						//fmt.Println("Its been 10 minutes")
						Alarm1.CurrentlyRunning = false
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm1.Sound && !Alarm1.Vibration {
				//fmt.Println("Sound")
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					time.Sleep(time.Second * 1)
					if !Alarm1.CurrentlyRunning {
						//fmt.Println("Breaking loop")
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						Alarm1.CurrentlyRunning = false
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}

			} else if !Alarm1.Sound && Alarm1.Vibration {
				fmt.Println("YOLO2")
				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm1.CurrentlyRunning {
							breaktime = true
							//fmt.Println("Breaking loop")
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						//fmt.Println("Its been ten minutes")
						Alarm1.CurrentlyRunning = false
						gpio.VibOff()
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm1.CurrentlyRunning = false
			}

		} else if Alarm2.Alarmtime == currenttime {
			NewIP = utils.GetIP()
			utils.Send(NewIP)
			Alarm2.CurrentlyRunning = true
			if Alarm2.Sound && Alarm2.Vibration {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm2.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm2.Sound && !Alarm2.Vibration {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					time.Sleep(time.Second * 1)
					if !Alarm2.CurrentlyRunning {
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm2.Sound && Alarm2.Vibration {
				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm2.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						gpio.VibOff()
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm2.CurrentlyRunning = false
			}

		} else if Alarm3.Alarmtime == currenttime {
			NewIP = utils.GetIP()
			utils.Send(NewIP)
			Alarm3.CurrentlyRunning = true
			if Alarm3.Sound && Alarm3.Vibration {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm3.Sound && !Alarm3.Vibration {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					time.Sleep(time.Second * 1)
					if !Alarm3.CurrentlyRunning {
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm3.Sound && Alarm3.Vibration {
				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						gpio.VibOff()
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm3.CurrentlyRunning = false
			}

		} else if Alarm4.Alarmtime == currenttime {
			NewIP = utils.GetIP()
			utils.Send(NewIP)
			Alarm4.CurrentlyRunning = true
			if Alarm4.Sound && Alarm4.Vibration {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}

			} else if Alarm4.Sound && !Alarm4.Vibration {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}
				for {
					time.Sleep(time.Second * 1)
					if !Alarm4.CurrentlyRunning {
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						break
					}
				}
			} else if !Alarm4.Sound && Alarm4.Vibration {
				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						gpio.VibOff()
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm4.CurrentlyRunning = false
			}
		}
	})
	c.Start()

	// Server index.html under ./public/index.html
	fs := http.FileServer(http.Dir(utils.Pwd() + "/public"))
	http.Handle("/", fs)

	//Handle the AJAX post call to submit a new time for a certain alarm
	http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		name := r.FormValue("name")
		time := r.FormValue("value")
		//fmt.Println(name)
		//Check to see which alarm the user is actually trying to modify, and modify the correct internally stored time
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
		//Write back the alarm data back to the ./public/json/alarms.json so Prometheus could retreive the data after a restart
		utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
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

		//Check to see which alarm the user is actually trying to modify, and modify the correct internally stored sound
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
		//Write back the alarm data back to the ./public/json/alarms.json so Prometheus could retreive the data after a restart
		utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
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
		//Check to see which alarm the user is actually trying to modify, and modify the correct internally stored vibration
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
		//Write back the alarm data back to the ./public/json/alarms.json so Prometheus could retreive the data after a restart
		utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
	})

	http.HandleFunc("/snooze", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("snoozed")
		//Using the AddTime() function, add 10 minutes to the currently running alarm, turn off the currently running alarm, and write back the correct configuration back to ./public/json/alarms.json
		if Alarm1.CurrentlyRunning {
			Alarm1.CurrentlyRunning = false
			Alarm1.AddTime(Alarm1.Alarmtime, "m", 10)
			utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
		} else if Alarm2.CurrentlyRunning {
			Alarm2.CurrentlyRunning = false
			Alarm2.AddTime(Alarm2.Alarmtime, "m", 10)
			utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
		} else if Alarm3.CurrentlyRunning {
			Alarm3.CurrentlyRunning = false
			Alarm3.AddTime(Alarm3.Alarmtime, "m", 10)
			utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
		} else if Alarm4.CurrentlyRunning {
			Alarm4.CurrentlyRunning = false
			Alarm4.AddTime(Alarm4.Alarmtime, "m", 10)
			utils.WriteBackJson(Alarm1, Alarm2, Alarm3, Alarm4, utils.Pwd()+"/public/json/alarms.json")
		}
		http.Redirect(w, r, "/", 301)
	})

	//Pass on the AJAX post /upload handler to the uploadHandler() function
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
