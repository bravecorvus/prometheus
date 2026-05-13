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
)

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

// ConvertBoolToOnOff renders a bool as the "on"/"off" strings used by the
// legacy /json/alarms.json read shape that the frontend still consumes.
func ConvertBoolToOnOff(arg bool) string {
	if arg {
		return "on"
	}
	return "off"
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

// padRGB zero-pads an RGB component to 3 digits (e.g. 5 -> "005", 42 -> "042").
func padRGB(val uint8) string {
	return fmt.Sprintf("%03d", val)
}

// ParseHexToRGB converts a hex color string to zero-padded RGB strings.
func ParseHexToRGB(hexStr string) (string, string, string) {
	hex, err := colors.ParseHEX(hexStr)
	if err != nil {
		fmt.Println("ERROR failed to parse hex color:", hexStr)
		return "000", "000", "000"
	}
	rgb := hex.ToRGB()
	return padRGB(rgb.R), padRGB(rgb.G), padRGB(rgb.B)
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
