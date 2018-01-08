// prometheus is the main program logic that runs both the alarm logic and the web server hosting the front-end user interface for users to control the Prometheus. The main function is split into two parts: First is the part that runs two cron jobs: the first runs once every second to try to send the time + LED string to the nixie clock and the second once every minute to check if the current time matches a user supplied alarm (also running the releavnt alarm actions vibrate and or output sound based on the user set parameters). Then the second half deals with providing the web server functionality. The fileserver serves a plain HTML file whose client side scripting uses Vue.js, Bootstrap, and jQuery to read from the JSON files stored at ./public/json/...json to display the information. In this way, the program avoids having to implement a database. The rest of the code provides HTTP POST endpoints when the user submits information. For most of these, the logic involves unmarshaling the HTTP headers, extracting the information, updating the program variables, and writing back the changes to the JSON files (so that the changes survive a program crash, and when the user reloads the UI, it will read from the updated settings.)

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

	"github.com/gilgameshskytrooper/prometheus/gpio"
	"github.com/gilgameshskytrooper/prometheus/nixie"
	"github.com/gilgameshskytrooper/prometheus/structs"
	"github.com/gilgameshskytrooper/prometheus/utils"
	"github.com/jacobsa/go-serial/serial"

	"github.com/robfig/cron"
)

var (
	//Alarm object from the structs.Alarm{} struct
	Alarm1 = structs.Alarm{}
	//Alarm object from the structs.Alarm{} struct
	Alarm2 = structs.Alarm{}
	//Alarm object from the structs.Alarm{} struct
	Alarm3 = structs.Alarm{}
	//Alarm object from the structs.Alarm{} struct
	Alarm4 = structs.Alarm{}
	// Bool to check if the Nixie Clock was found.
	foundNixie bool
	// Bool to keep track of whether the user wants to receive emails when the Pi's IP changes.
	// (If using DDNS, this is not necessary as the DDNS client will automatically update the relevant IP, and it will always be reachable with the same command)
	EnableEmail bool
	// The email address to send the IP change notification email to
	Email string
	// Bool to keep track of whether shairport-sync is installed
	shairportInstalled bool
	// Alarm sound filename
	Soundname string
	// Used to see if user wants to use the cvlc command for use on a custom sound card, or the regular cvlc command
	CustomSoundCard bool
	// Used to see if user wants to enable LED's or not
	EnableLed bool
	// Used to tell what colors the user wants the LED to be on the clock
	Red, Green, Blue string
	Options          = serial.OpenOptions{}
	Port             io.ReadWriteCloser
)

//General error handler: I guess it wasn't used nearly as much as should to warrant it's existance, but its here nonetheless
func Errhandler(err error) {
	if err != nil {
		fmt.Println("ERROR")
	}
}

//Function to check whether the alarm has been running for more than 10 minutes
func OverTenMinutes(alarmtime string) bool {
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

	timecurrent := time.Now()
	difference := timecurrent.Minute() - time.Date(int(year), month, int(day), hour, minutes, 0, 0, time.Local).Minute()
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
		log.Fatalln(err)
	}

	// If a new file got uploaded, make sure this gets reflected in the program var Soundname and also write out the new name into ./public/json/trackinfo
	Soundname = strings.TrimSpace(b.String())
	d1 := []byte(Soundname)
	errrrrrrrrr := ioutil.WriteFile(utils.Pwd()+"/public/json/trackinfo", d1, 0644)
	if errrrrrrrrr != nil {
		fmt.Println(errrrrrrrrr)
	}

	fmt.Fprintf(w, header.Filename)
}

//Initialization sequence
// 1. Grab the persistent alarm information from the alarms.json file
// 2. Use this data to store legitimate values of time, vibration, and sound into the 4 struct.Alarm objects
// 3. Grab the name of the current sound file via an "ls" shell command, and save it to the global variable Sound, and then write that information into ./public/json/trackinfo
// 4. Get the email of user to be used in the CheckIPChange() function
// 5. Get the persistent data of whether or not the user wants Prometheus to send emails when the IP changes. (Note, this is probably alot easier done through a dynamic DNS service which runs a background program to constanty check the IP of the Pi, updates a domain name server, and then you can access that as a link such as myclockname.ddns.net:3000
func init() {
	//Save the JSON alarms configurations into the mold
	jsondata := structs.GetRawJson(utils.Pwd() + "/public/json/alarms.json")
	Alarm1.InitializeAlarms(jsondata, 0)
	Alarm2.InitializeAlarms(jsondata, 1)
	Alarm3.InitializeAlarms(jsondata, 2)
	Alarm4.InitializeAlarms(jsondata, 3)

	var b bytes.Buffer
	if err := utils.Execute(&b,
		exec.Command("ls", utils.Pwd()+"/public/assets"),
	); err != nil {
		log.Fatalln(err)
	}

	Soundname = strings.TrimSpace(b.String())
	d1 := []byte(Soundname)
	errrrrrrrrr := ioutil.WriteFile(utils.Pwd()+"/public/json/trackinfo", d1, 0644)
	if errrrrrrrrr != nil {
		fmt.Println(errrrrrrrrr)
	}
	Email = utils.GetEmail()
	EnableEmail = utils.GetEnableEmail()
	CustomSoundCard = utils.UseCustomSoundCard()
	Red, Green, Blue, EnableLed = utils.ColorInitialize()
	Options.PortName = nixie.FindArduino()
	Options.BaudRate = 115200
	Options.DataBits = 8
	Options.StopBits = 1
	Options.MinimumReadSize = 4
}

// Main function
// Runs the cron job (checking once a minute at exactly the point when second is 00) to check if the current time matches the user supplied alarm time configuration, and then runs the alarm if an enabled alarm matches the time
// Runs a separate cron job (once every second) to send the current time as a string to the Nixie Clock through serial USB
// Also, main contains all the http HandleFunc's to deal with GET '/', POST '/time', POST '/sound', POST '/vibration', POST '/snooze', POST '/enableemail', POST '/newemail'
func main() {

	// Ensure any previous incarnations of shairport-sync gets killed.
	// if no previous process exists, KillShairportSync() automatically handles this.
	shairportInstalled = utils.CheckShairportSyncInstalled()
	if shairportInstalled {
		shairportstart := exec.Command("shairport-sync", "-d")
		shairportstarterror := shairportstart.Run()
		if shairportstarterror != nil {
			fmt.Println("Could not start shairport-sync daemon")
		}
	}
	// Options = serial.OpenOptions{
	// PortName:        nixie.FindArduino(),
	// BaudRate:        115200,
	// DataBits:        8,
	// StopBits:        1,
	// MinimumReadSize: 4,
	// }

	// Open the serial USB port to communicate with the clock.
	if Options.PortName != "" {
		Port, err := serial.Open(Options)
		if err != nil {
			foundNixie = false
		} else {
			foundNixie = true
		}
		defer Port.Close()
	}

	// Make sure to close it later.

	// Initialize all 4 instances of alarm clocks
	// Create function that updates clock once a minute (used to see if any times match up)
	t := time.Now()
	currenttime := t.Format("15:04")
	c := cron.New()

	// Send relevant time clock over serial USB
	c.AddFunc("@every 1s", func() {
		// fmt.Println("RGB(", Red, Green, Blue, ")")

		if EnableLed {
			if foundNixie {
				b := []byte(nixie.CurrentTimeAsString() + Red + Green + Blue)
				_, err := Port.Write(b)
				if err != nil {
					log.Fatalf("Port.Write: %v", err)
				}
			} else {
				Options.PortName = nixie.FindArduino()
				if Options.PortName != "" {
					foundNixie = true
				} else {
					foundNixie = false
				}
			}

		} else {

			if foundNixie {
				b := []byte(nixie.CurrentTimeAsString())
				_, err := Port.Write(b)
				if err != nil {
					log.Fatalf("Port.Write: %v", err)
				}
			} else {
				Options.PortName = nixie.FindArduino()
				if Options.PortName != "" {
					foundNixie = true
				} else {
					foundNixie = false
				}
			}
		}

	})

	//Run the following once a minute
	//Check all 4 alarms to see if the current time matches any configurations
	c.AddFunc("0 * * * * *", func() {
		breaktime := false
		duration := time.Second * 3
		t = time.Now()
		currenttime = t.Format("15:04")
		if EnableEmail {
			utils.CheckIPChange()
		}

		if Alarm1.Alarmtime == currenttime {

			go utils.RestartNetwork()
			Alarm1.CurrentlyRunning = true

			if Alarm1.Sound && Alarm1.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {
					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
					errrrror := playsound.Start()
					if errrrror != nil {
						fmt.Println("ERRRRRROR")
					}

					for {
						gpio.VibOn()
						for i := 1; i <= 50; i++ {
							time.Sleep(time.Millisecond * 50)
							if !Alarm1.CurrentlyRunning {
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm1.Alarmtime) {
							Alarm1.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}
				} else {
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm1.Alarmtime) {
							Alarm1.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}
				}

			} else if Alarm1.Sound && !Alarm1.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"
				if shairportInstalled {
					utils.KillShairportSync()
				}
				if CustomSoundCard {
					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
					errrrror := playsound.Start()
					if errrrror != nil {
						fmt.Println("ERRRRRROR")
					}
					for {
						time.Sleep(time.Second * 1)
						if !Alarm1.CurrentlyRunning {
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm1.Alarmtime) {
							Alarm1.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						}
					}
				} else {

					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname)
					errrrror := playsound.Start()
					if errrrror != nil {
						fmt.Println("ERRRRRROR")
					}
					for {
						time.Sleep(time.Second * 1)
						if !Alarm1.CurrentlyRunning {
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm1.Alarmtime) {
							Alarm1.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}
				}

			} else if !Alarm1.Sound && Alarm1.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm1.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						breaktime = false
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break

					} else if OverTenMinutes(Alarm1.Alarmtime) {
						Alarm1.CurrentlyRunning = false
						gpio.VibOff()
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm1.CurrentlyRunning = false
			}

		} else if Alarm2.Alarmtime == currenttime {

			// Check if there is network connectivity (if not, then restart network interfaces)
			go utils.RestartNetwork()
			Alarm2.CurrentlyRunning = true

			if Alarm2.Sound && Alarm2.Vibration {
				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {

					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
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
								Red, Green, Blue, EnableLed = utils.ColorInitialize()
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm2.Alarmtime) {
							Alarm2.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}
				} else {

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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm2.Alarmtime) {
							Alarm2.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}
				}

			} else if Alarm2.Sound && !Alarm2.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {

					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm2.Alarmtime) {
							Alarm2.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}

				} else {

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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm2.Alarmtime) {
							Alarm2.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						}
					}
				}

			} else if !Alarm2.Sound && Alarm2.Vibration {
				Red = "255"
				Green = "000"
				Blue = "000"

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
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break

					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						gpio.VibOff()
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm2.CurrentlyRunning = false
			}

		} else if Alarm3.Alarmtime == currenttime {
			// Check if there is network connectivity (if not, then restart network interfaces)
			go utils.RestartNetwork()
			Alarm3.CurrentlyRunning = true

			if Alarm3.Sound && Alarm3.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {
					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm3.Alarmtime) {
							Alarm3.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}
				} else {
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}

							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm3.Alarmtime) {
							Alarm3.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}
				}

			} else if Alarm3.Sound && !Alarm3.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {
					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()

							break
						} else if OverTenMinutes(Alarm3.Alarmtime) {
							Alarm3.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}

				} else {
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm3.Alarmtime) {
							Alarm3.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}

				}

			} else if !Alarm3.Sound && Alarm3.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

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
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						gpio.VibOff()
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}
				}
			} else {
				Alarm3.CurrentlyRunning = false
			}

		} else if Alarm4.Alarmtime == currenttime {
			// Check if there is network connectivity (if not, then restart network interfaces)
			go utils.RestartNetwork()
			Alarm4.CurrentlyRunning = true

			if Alarm4.Sound && Alarm4.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {
					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm4.Alarmtime) {
							Alarm4.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}

				} else {
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm4.Alarmtime) {
							Alarm4.CurrentlyRunning = false
							gpio.VibOff()
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else {
							gpio.VibOff()
							time.Sleep(duration)
						}

					}

				}

			} else if Alarm4.Sound && !Alarm4.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				if shairportInstalled {
					utils.KillShairportSync()
				}

				if CustomSoundCard {
					var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+Soundname, "-A=alsa", "--alsa-audio-device=default")
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break

						} else if OverTenMinutes(Alarm4.Alarmtime) {
							Alarm4.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}

				} else {
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
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						} else if OverTenMinutes(Alarm4.Alarmtime) {
							Alarm4.CurrentlyRunning = false
							errrrrorkill := playsound.Process.Kill()
							if errrrrorkill != nil {
								fmt.Println("ERRRRRROR")
							}
							if shairportInstalled {
								shairportdaemon := exec.Command("shairport-sync", "-d")
								shairportdaemonerror := shairportdaemon.Run()
								if shairportdaemonerror != nil {
									fmt.Println("Could not start shairport-sync daemon")
								}
							}
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}

				}

			} else if !Alarm4.Sound && Alarm4.Vibration {

				Red = "255"
				Green = "000"
				Blue = "000"

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm4.CurrentlyRunning {
							breaktime = true
							Red, Green, Blue, EnableLed = utils.ColorInitialize()
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						breaktime = false
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						gpio.VibOff()
						Red, Green, Blue, EnableLed = utils.ColorInitialize()
						break
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

	http.HandleFunc("/enableemail", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		value := r.FormValue("value")
		if value == "true" {
			EnableEmail = true
			utils.WriteEnableEmail("true")
		} else {
			EnableEmail = false
			utils.WriteEnableEmail("false")
		}
	})

	http.HandleFunc("/customsoundcard", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		value := r.FormValue("value")
		if value == "true" {
			EnableEmail = true
			utils.WriteCustomSoundCard("true")
		} else {
			EnableEmail = false
			utils.WriteCustomSoundCard("false")
		}
	})

	http.HandleFunc("/newemail", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		value := r.FormValue("value")
		utils.WriteEmail(value)
	})

	http.HandleFunc("/submitcolors", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		value := r.FormValue("value")
		Red, Green, Blue = utils.ColorUpdate(value)
	})

	http.HandleFunc("/submitenableled", func(w http.ResponseWriter, r *http.Request) {
		erawr := r.ParseForm()
		if erawr != nil {
			fmt.Println("ERROR")
			os.Exit(1)
		}
		value := r.FormValue("value")
		if value == "true" {
			EnableLed = true
			utils.WriteEnableLed("true")
		} else {
			EnableLed = false
			utils.WriteEnableLed("false")
		}
	})

	//Pass on the AJAX post /upload handler to the uploadHandler() function
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Listening...")
	if utils.Exists(utils.Pwd()+"/server.crt") && utils.Exists(utils.Pwd()+"/server.key") {
		log.Fatal(http.ListenAndServeTLS(":3000", utils.Pwd()+"/server.crt", utils.Pwd()+"/server.key", nil))
	} else {
		fmt.Println("If you want the program to utilize TLS (i.e. host an encrypted HTTPS front end, please do the following in command line in the same directory as the bigdisk executable to first create a private self-signed rsa key, then a public key (x509) key based on the private key:\n\topenssl genrsa -out server.key 2048\n\topenssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650\nThen make sure you finish filling in the details asked in command line.\n\nFor now, unencrypted http will be used.")
		log.Fatal(http.ListenAndServe(":3000", nil))
	}

}
