package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/go-playground/colors.v1"

	"prometheus/structs"
)

func WriteIP(arg string) {
	if err := os.WriteFile(Pwd()+"/public/json/ip", []byte(arg), 0644); err != nil {
		fmt.Println("ERROR WriteIP()")
	}
}

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
		fmt.Println(string(error_buffer.Bytes()), err)
	}
	return err
}

func GetIP() string {
	var b bytes.Buffer
	if err := Execute(&b,
		exec.Command("ifconfig", "wlan0"),
		exec.Command("grep", "inet"),
		exec.Command("awk", "NR==1{print $2}"),
	); err != nil {
		fmt.Println(err)
	}
	str := b.String()
	regex, _ := regexp.Compile("\n")
	str = regex.ReplaceAllString(str, "")
	return strings.TrimSpace(str)
}

func GetIPFromFile() string {
	content, err := os.ReadFile(Pwd() + "/public/json/ip")
	if err != nil {
		fmt.Println("ERROR GetIPFromFile()")
	}
	return strings.Split(string(content), "\n")[0]
}

func UseCustomSoundCard() bool {
	content, err := os.ReadFile(Pwd() + "/public/json/customsoundcard")
	if err != nil {
		fmt.Println("ERROR UseCustomSoundCard()")
	}
	return strings.Split(string(content), "\n")[0] == "true"
}

func GetEnableEmail() bool {
	content, err := os.ReadFile(Pwd() + "/public/json/enableemail")
	if err != nil {
		fmt.Println("ERROR GetEnableEmail()")
	}
	return strings.Split(string(content), "\n")[0] == "True"
}

func GetEmail() string {
	content, err := os.ReadFile(Pwd() + "/public/json/email")
	if err != nil {
		fmt.Println("ERROR GetEmail()")
	}
	return strings.Split(string(content), "\n")[0]
}

func CheckIPChange() {
	if GetIPFromFile() != GetIP() {
		WriteIP(GetIP())
		sendemail := exec.Command("email/prometheusemail", GetEmail(), GetIP())
		if err := sendemail.Run(); err != nil {
			fmt.Println("failed to send email")
		}
	}
}

func convertBooltoString(arg bool) string {
	if arg {
		return "on"
	}
	return "off"
}

// WriteBackJson writes alarm configurations to the JSON file.
// Now accepts a slice instead of 4 separate arguments.
func WriteBackJson(alarms []structs.Alarm, filepath string) {
	var parts []string
	for _, a := range alarms {
		parts = append(parts, fmt.Sprintf(
			`{"name":"%s","time":"%s","sound":"%s","vibration":"%s"}`,
			a.Name, a.Alarmtime, convertBooltoString(a.Sound), convertBooltoString(a.Vibration),
		))
	}
	content := "[" + strings.Join(parts, ",\n") + "]"
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		fmt.Println("Error writing back JSON alarm file for " + filepath)
	}
}

func Pwd() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	return dir
}

func RestartNetwork() {
	_, err := http.Get("http://google.com")
	if err != nil {
		ifdown := exec.Command("ifdown", "wlan0")
		if err := ifdown.Run(); err != nil {
			fmt.Println("ifdown wlan0 command failed")
		}
		time.Sleep(time.Second * 5)
		ifup := exec.Command("ifup", "wlan0")
		if err := ifup.Run(); err != nil {
			fmt.Println("ifup wlan0 command failed")
			go RestartNetwork()
		}
	}
}

func WriteEnableEmail(arg string) {
	if err := os.WriteFile(Pwd()+"/public/json/enableemail", []byte(arg), 0644); err != nil {
		fmt.Println("Error writing back enableemail file")
	}
}

func WriteEmail(arg string) {
	if err := os.WriteFile(Pwd()+"/public/json/email", []byte(arg), 0644); err != nil {
		fmt.Println("Error writing back email file")
	}
}

func CheckShairportRunning() bool {
	var b bytes.Buffer
	if err := Execute(&b,
		exec.Command("ps", "aux"),
		exec.Command("grep", "shairport"),
		exec.Command("awk", "NR==1{print $NF}"),
	); err != nil {
		fmt.Println(err.Error())
	}
	return strings.TrimSpace(b.String()) != "shair"
}

func KillShairportSync() {
	if CheckShairportRunning() {
		var b bytes.Buffer
		if err := Execute(&b,
			exec.Command("ps", "aux"),
			exec.Command("grep", "shairport"),
			exec.Command("awk", "NR==1{print $2}"),
		); err != nil {
			fmt.Println(err)
		}
		if err := exec.Command("kill", strings.TrimSpace(b.String())).Run(); err != nil {
			fmt.Println("Could not kill shairport-sync")
		}
	}
}

func CheckShairportSyncInstalled() bool {
	var stdout bytes.Buffer
	cmd := exec.Command("which", "shairport-sync")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("which shairport-sync command failed")
	}
	return strings.TrimSpace(stdout.String()) != ""
}

func WriteCustomSoundCard(arg string) {
	if err := os.WriteFile(Pwd()+"/public/json/customsoundcard", []byte(arg), 0644); err != nil {
		fmt.Println("Error writing back customsoundcard file")
	}
}

// padRGB zero-pads an RGB component to 3 digits (e.g. 5 -> "005", 42 -> "042").
func padRGB(val uint8) string {
	return fmt.Sprintf("%03d", val)
}

// parseHexToRGB converts a hex color string to zero-padded RGB strings.
func parseHexToRGB(hexStr string) (string, string, string) {
	hex, err := colors.ParseHEX(hexStr)
	if err != nil {
		fmt.Println("ERROR failed to parse hex color:", hexStr)
		return "000", "000", "000"
	}
	rgb := hex.ToRGB()
	return padRGB(rgb.R), padRGB(rgb.G), padRGB(rgb.B)
}

func ColorUpdate(arg string) (string, string, string) {
	if err := os.WriteFile(Pwd()+"/public/json/colors", []byte(arg), 0644); err != nil {
		fmt.Println("Error writing back colors file")
	}
	return parseHexToRGB(arg)
}

func ColorInitialize() (string, string, string, bool) {
	content, err := os.ReadFile(Pwd() + "/public/json/colors")
	if err != nil {
		fmt.Println("ERROR ColorInitialize() read colors JSON file")
	}
	hexStr := strings.Split(string(content), "\n")[0]
	r, g, b := parseHexToRGB(hexStr)

	ledContent, err := os.ReadFile(Pwd() + "/public/json/enableled")
	if err != nil {
		fmt.Println("ERROR ColorInitialize() read enableled file")
	}
	enableLed := strings.Split(string(ledContent), "\n")[0] == "true"

	return r, g, b, enableLed
}

func WriteEnableLed(arg string) {
	if err := os.WriteFile(Pwd()+"/public/json/enableled", []byte(arg), 0644); err != nil {
		fmt.Println("Error writing back enableled file")
	}
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func OverTenMinutes(alarmtime string) bool {
	year, month, day := time.Now().Date()
	parts := strings.Split(alarmtime, ":")
	if len(parts) != 2 {
		return false
	}
	hour, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])

	alarmTime := time.Date(year, month, day, hour, minutes, 0, 0, time.Local)
	diff := time.Since(alarmTime).Minutes()
	return diff >= 10
}
