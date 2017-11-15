package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"../structs"
)

//Declare the identity of the current wlan0 IP. Used to check if the IP changed.
var IP, NewIP string

//Taking in the IP as a string as the argument, write the IP address to ./public/json/ip to use when the program is restarted
func WriteIP(arg string) {
	writebuf := []byte(arg)
	err := ioutil.WriteFile(Pwd()+"/public/json/ip", writebuf, 0644)
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

	if err := Call(stack, pipe_stack); err != nil {
		log.Fatalln(string(error_buffer.Bytes()), err)
	}
	return err
}

//Used to execute complex pipes to filter out the wlan0 IP address of the Pi via ifconfig, awk, and cut
func Call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
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
				err = Call(stack[1:], pipes[1:])
			}
		}()
	}
	return stack[0].Wait()
}

//Function that returns the current wlan0 address as a string
func GetIP() string {
	var b bytes.Buffer
	var str string
	if err := Execute(&b,
		//Since piping commands are a bit of a pain, using the above functions Call() and Execute(), execute "/sbin/ifconfig wlan0 | grep 'inet addr:' | cut -d -f2 | awk '{print $1}'"
		exec.Command("ifconfig", "wlan0"),
		exec.Command("grep", "inet"),
		exec.Command("awk", "NR==1{print $2}"),
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
func GetIPFromFile() string {
	content, err := ioutil.ReadFile(Pwd() + "/public/json/ip")
	if err != nil {
		fmt.Println("ERROR")
	}
	lines := strings.Split(string(content), "\n")
	//fmt.Println("Get IP From File", lines[0])
	return lines[0]
}

//grab the email from "./public/json/email" to be used if the user has a dynamically assigned IP, and the IP changes from before
func GetEmail() string {
	content, err := ioutil.ReadFile(Pwd() + "/public/json/email")
	if err != nil {
		fmt.Println("ERROR")
	}
	lines := strings.Split(string(content), "\n")
	return (lines[0])
}

//Function that checks to see if the current IP matches the IP string currently registered.
//If the old IP and the new IP don't match, send the user an email notifying them of this change. Please change the stored at ./public/json/ip to get these notifications
func Send(body string) {
	if body == IP {
		//If the IP didn't change, just ignore
		return
	} else {
		IP = NewIP
		WriteIP(IP)
		//Account from which Prometheus sends an email from.
		from := "email@example.com"
		pass := "password"
		var to string
		to = GetEmail()

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

//Since the Sound and Vibration variables are stored as "on" or "off" in the alarms.json file, this function converts a boolean to the on/off format
func convertBooltoString(arg bool) string {
	if arg {
		return "on"
	} else {
		return "off"
	}
}

//Write back the correct alarm configurations to ./public/json/alarms.json so that the information can be retrieved when ./main is restarted
func WriteBackJson(Alarm1 structs.Alarm, Alarm2 structs.Alarm, Alarm3 structs.Alarm, Alarm4 structs.Alarm, filepath string) {
	content := []byte("[{\"name\":\"" + Alarm1.Name + "\",\"time\":\"" + Alarm1.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm1.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm1.Vibration) + "\"},\n{\"name\":\"" + Alarm2.Name + "\",\"time\":\"" + Alarm2.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm2.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm2.Vibration) + "\"},\n{\"name\":\"" + Alarm3.Name + "\",\"time\":\"" + Alarm3.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm3.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm3.Vibration) + "\"},\n{\"name\":\"" + Alarm4.Name + "\",\"time\":\"" + Alarm4.Alarmtime + "\",\"sound\":\"" + convertBooltoString(Alarm4.Sound) + "\",\"vibration\":\"" + convertBooltoString(Alarm4.Vibration) + "\"}]")
	err := ioutil.WriteFile(filepath, content, 0644)
	if err != nil {
		fmt.Println("Error writing back JSON alarm file for " + filepath)
		os.Exit(1)
	}
}

func Pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

// RestartIfNoIP is a function that restarts the Network if not network is detected
func RestartNetwork() {
	_, err := http.Get("http://google.com")
	if err != nil {
		var b bytes.Buffer
		if err := Execute(&b,
			exec.Command("ifdown", "wlan0"),
		); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second * 5)
		if err := Execute(&b,
			exec.Command("ifup", "--force", "wlan0"),
		); err != nil {
			log.Fatalln(err)
		}
	}
}
