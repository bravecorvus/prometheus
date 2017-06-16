//main.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robfig/cron"
	"github.com/stianeikeland/go-rpio"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//JSON struct used when reading in the ./public/json/alarms.json to get the alarm configurations when main is first started
type jsonAlarms struct {
	JsonName      string `json:"name"`
	JsonTime      string `json:"time"`
	JsonSound     string `json:"sound"`
	JsonVibration string `json:"vibration"`
}

//How Prometheus stores the Alarms internally. all 4 alarms are of type "Alarm"
type Alarm struct {
	Name             string //Name of the alarm: one of the following: alarm1, alarm2, alarm3, or alarm4
	Alarmtime        string //Time of the alarm, stored as time; in 24 hour time (e.g. 11:00PM = 23:00).
	Sound            bool   //Sets whether Prometheus should play sound or not
	Vibration        bool   //Sets whether Prometheus should vibrate or not
	CurrentlyRunning bool   //Set to true when alarm is running, false when the user tells it to stop (i.e. Sound/Vibration toggle is turned off for current alarm, it has been more then 10 minutes, or the user pressed the snooze button)
}

//Declare the alarms
var Alarm1 = Alarm{}
var Alarm2 = Alarm{}
var Alarm3 = Alarm{}
var Alarm4 = Alarm{}

//Declare the name of the alarm sound stored in ./public/assets/sound_name.extension
var Soundname string

//Declare the identity of the current wlan0 IP. Used to check if the IP changed.
var IP, NewIP string

//Taking in the IP as a string as the argument, write the IP address to ./public/json/ip to use when the program is restarted
func writeIP(arg string) {
	writebuf := []byte(arg)
	err := ioutil.WriteFile("./public/json/ip", writebuf, 0644)
	if err != nil {
		fmt.Println("ERROR")
	}
}

//Used to execute complex pipes to filter out the wlan0 IP address of the Pi via ifconfig, awk, and cut
func Execute(output_buffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
	var error_buffer bytes.Buffer
	pipe_stack := make([]*io.PipeWriter, len(stack)-1)
	i := 0
	for ; i < len(stack)-1; i++ {
		stdin_pipe, stdout_pipe := io.Pipe()
		stack[i].Stdout = stdout_pipe
		stack[i].Stderr = &error_buffer
		stack[i+1].Stdin = stdin_pipe
		pipe_stack[i] = stdout_pipe
	}
	stack[i].Stdout = output_buffer
	stack[i].Stderr = &error_buffer

	if err := call(stack, pipe_stack); err != nil {
		log.Fatalln(string(error_buffer.Bytes()), err)
	}
	return err
}

//Used to execute complex pipes to filter out the wlan0 IP address of the Pi via ifconfig, awk, and cut
func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}
	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				pipes[0].Close()
				err = call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}

//Function that returns the current wlan0 address as a string
func getIP() string {
	var b bytes.Buffer
	var str string
	if err := Execute(&b,
		//Since piping commands are a bit of a pain, using the above functions call() and Execute(), execute "/sbin/ifconfig wlan0 | grep 'inet addr:' | cut -d -f2 | awk '{print $1}'"
		exec.Command("/sbin/ifconfig", "wlan0"),
		exec.Command("grep", "inet addr:"),
		exec.Command("cut", "-d:", "-f2"),
		exec.Command("awk", "{print $1}"),
	); err != nil {
		log.Fatalln(err)
	}
	str = b.String()
	regex, err := regexp.Compile("\n")
	if err != nil {
		fmt.Println("ERROR")
	}
	str = regex.ReplaceAllString(str, "")
	//fmt.Println("Get IP", str)
	return strings.TrimSpace(str)
}

//Read the IP from the file, "./public/json/ip", return it as a string
func getIPFromFile() string {
	content, err := ioutil.ReadFile("./public/json/ip")
	if err != nil {
		fmt.Println("ERROR")
	}
	lines := strings.Split(string(content), "\n")
	//fmt.Println("Get IP From File", lines[0])
	return lines[0]
}

//grab the email from "./public/json/email" to be used if the user has a dynamically assigned IP, and the IP changes from before
func getEmail() string {
	content, err := ioutil.ReadFile("./public/json/email")
	if err != nil {
		fmt.Println("ERROR")
	}
	lines := strings.Split(string(content), "\n")
	return (lines[0])
}

//Function that checks to see if the current IP matches the IP string currently registered.
//If the old IP and the new IP don't match, send the user an email notifying them of this change. Please change the stored at ./public/json/ip to get these notifications
func send(body string) {
	if body == IP {
		return
	} else {
		IP = NewIP
		writeIP(IP)
		//Account from which Prometheus sends an email from.
		from := "email@example.com"
		pass := "password"
		var to string
		to = getEmail()

		msg := "From: " + from + "\n" +
			"To: " + to + "\n" +
			"Subject: New Prometheus IP: " +
			body

		err := smtp.SendMail("smtp.gmail.com:587",
			smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
			from, []string{to}, []byte(msg))

		if err != nil {
			log.Printf("smtp error: %s", err)
			return
		}

		log.Print("sent")
	}
}

//Sends the signal to turn on the bed vibrator by sending a High (true) signal to GPIO 17
func VibOn() {
	// fmt.Println("VibOn")
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

//Sends the signal to turn off the bed vibrator by sending a Low (false) signal to GPIO 17
func VibOff() {
	// fmt.Println("VibOff")
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

//Get the alarms configurations from the ./public/json/alarms.json
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

//Populate the values of all 4 internally stored alarms using the values stored in the ./public/json/alarms.json
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

//General error handler. I guess it wasn't used nearly as much as should to warrant it's existance, but its here nonetheless.
func Errhandler(err error) {
	if err != nil {
		fmt.Println("ERROR")
	}
}

//Function that adds 10 minutes to the currently running alarm. Necessary because internally, alarm times are stored as strings, rather than the time class
func (arg *Alarm) addTime(originaltime string, hms string, byhowmuch int) { //takes originaltime, and adds byhowmuch hours/minutes/seconds, then returns the string
	currenttime, _ := time.Parse("15:04", originaltime)
	//fmt.Println("before fixed snooze time", currenttime)
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
	//fmt.Println("fixed snooze time", arg.Alarmtime)
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

//Since the Sound and Vibration variables are stored as "on" or "off" in the alarms.json file, this function converts a boolean to the on/off format
func convertBooltoString(arg bool) string {
	if arg {
		return "on"
	} else {
		return "off"
	}
}

//Write back the correct alarm configurations to ./public/json/alarms.json so that the information can be retrieved when ./main is restarted
func writeBackJson(Alarm1 Alarm, Alarm2 Alarm, Alarm3 Alarm, Alarm4 Alarm, filepath string) {
	// fmt.Println("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"}\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"}\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"}\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	content := []byte("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"},\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"},\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"},\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	err := ioutil.WriteFile(filepath, content, 0644)
	if err != nil {
		fmt.Println("Error writing back JSON alarm file for " + filepath)
		os.Exit(1)
	}
}

//Handle the AJAX file upload from the client. Deletes the old file, and saves the new name of the sound from the header.Filename
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	//Delete the old alarm sound via shell command process rm public/assets/sound_name.extension
	rmerror := exec.Command("rm", "public/assets/"+Soundname).Run()
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
	out, err1 := os.Create("./public/assets/" + header.Filename)

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
	//Save the JSON alarms configurations into the mold
	jsondata := getRawJson("./public/json/alarms.json")
	//Set up the intial values for the alarms using the values we get from above
	Alarm1.initializeAlarms(jsondata, 0)
	Alarm2.initializeAlarms(jsondata, 1)
	Alarm3.initializeAlarms(jsondata, 2)
	Alarm4.initializeAlarms(jsondata, 3)
	//Read the most recent IP address from ./public/json/ip
	IP = strings.TrimSpace(getIPFromFile())
	//Grab the name of the alarm sound file via ls command to the ./public/assets/ folder.
	var b bytes.Buffer
	if err := Execute(&b,
		exec.Command("ls", "public/assets"),
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
			NewIP = getIP()
			send(NewIP)
			//fmt.Println("Alarm 1")
			Alarm1.CurrentlyRunning = true
			if Alarm1.Sound && Alarm1.Vibration {
				//fmt.Println("Sound and Vibration")
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
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
						VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						//fmt.Println("Its been 10 minutes")
						Alarm1.CurrentlyRunning = false
						VibOff()
						errrrrorkill := playsound.Process.Kill()
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
				//fmt.Println("Sound")
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
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
				for {
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm1.CurrentlyRunning {
							breaktime = true
							//fmt.Println("Breaking loop")
							break
						}
					}
					if breaktime {
						VibOff()
						breaktime = false
						break
					} else if OverTenMinutes(Alarm1.Alarmtime) {
						//fmt.Println("Its been ten minutes")
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
			NewIP = getIP()
			send(NewIP)
			Alarm2.CurrentlyRunning = true
			if Alarm2.Sound && Alarm2.Vibration {
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm2.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm2.Alarmtime) {
						Alarm2.CurrentlyRunning = false
						VibOff()
						errrrrorkill := playsound.Process.Kill()
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
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
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
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
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
			NewIP = getIP()
			send(NewIP)
			Alarm3.CurrentlyRunning = true
			if Alarm3.Sound && Alarm3.Vibration {
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm3.Alarmtime) {
						Alarm3.CurrentlyRunning = false
						VibOff()
						errrrrorkill := playsound.Process.Kill()
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
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
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
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
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
			NewIP = getIP()
			send(NewIP)
			Alarm4.CurrentlyRunning = true
			if Alarm4.Sound && Alarm4.Vibration {
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
				errrrror := playsound.Start()
				if errrrror != nil {
					fmt.Println("ERRRRRROR")
				}

				for {
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println("ERRRRRROR")
						}
						breaktime = false
						break
					} else if OverTenMinutes(Alarm4.Alarmtime) {
						Alarm4.CurrentlyRunning = false
						VibOff()
						errrrrorkill := playsound.Process.Kill()
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
				var playsound = exec.Command("cvlc", "./public/assets/"+Soundname)
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
					VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
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
		writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
	})

	http.HandleFunc("/snooze", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("snoozed")
		//Using the addTime() function, add 10 minutes to the currently running alarm, turn off the currently running alarm, and write back the correct configuration back to ./public/json/alarms.json
		if Alarm1.CurrentlyRunning {
			Alarm1.CurrentlyRunning = false
			Alarm1.addTime(Alarm1.Alarmtime, "m", 10)
			writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
		} else if Alarm2.CurrentlyRunning {
			Alarm2.CurrentlyRunning = false
			Alarm2.addTime(Alarm2.Alarmtime, "m", 10)
			writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
		} else if Alarm3.CurrentlyRunning {
			Alarm3.CurrentlyRunning = false
			Alarm3.addTime(Alarm3.Alarmtime, "m", 10)
			writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
		} else if Alarm4.CurrentlyRunning {
			Alarm4.CurrentlyRunning = false
			Alarm4.addTime(Alarm4.Alarmtime, "m", 10)
			writeBackJson(Alarm1, Alarm2, Alarm3, Alarm4, "./public/json/alarms.json")
		}
		http.Redirect(w, r, "/", 301)
	})

	//Pass on the AJAX post /upload handler to the uploadHandler() function
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
