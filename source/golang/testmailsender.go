package main

import (
	"bytes"
	"fmt"
	"github.com/robfig/cron"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"regexp"
	//"os"
	"os/exec"
	"strings"
)

var IP string

func init() {
	IP = getIPFromFile()
}
func main() {
	c := cron.New()
	c.AddFunc("0 * * * * *", func() {
		//fmt.Println(IP)
		newIP := getIP()
		if newIP == IP {
			send(newIP)
			IP = newIP
			writeIP(IP)
		}
		fmt.Println(newIP == IP)
		//fmt.Println(IP)
		//fmt.Println(newIP)
		//fmt.Println(IP)
	})
}

func writeIP(arg string) {
	writebuf := []byte(arg)
	err := ioutil.WriteFile("./public/json/ip", writebuf, 0644)
	if err != nil {
		fmt.Println("ERROR")
	}
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

	if err := call(stack, pipe_stack); err != nil {
		log.Fatalln(string(error_buffer.Bytes()), err)
	}
	return err
}

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

func getIP() string {
	var b bytes.Buffer
	var str string
	if err := Execute(&b,
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
	return str
}

func getIPFromFile() string {
	content, err := ioutil.ReadFile("./public/json/ip")
	if err != nil {
		fmt.Println("ERROR")
	}
	lines := strings.Split(string(content), "\n")
	//fmt.Println("Get IP From File", lines[0])
	return lines[0]
}

func getEmail() string {
	content, err := ioutil.ReadFile("./public/json/email")
	if err != nil {
		fmt.Println("ERROR")
	}
	lines := strings.Split(string(content), "\n")
	return (lines[0])
}

func send(body string) {
	from := "prometheusclock@gmail.com"
	pass := "abcprometheusclock"
	var to string
	to = getEmail()

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: IP Change from Prometheus " +
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
