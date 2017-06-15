package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		var Playsound = exec.Command("afplay", "./public/assets/01 K..mp3")
		errrrror := Playsound.Start()
		if errrrror != nil {
			fmt.Println("ERROR")
		}
		time.Sleep(time.Second * 10)
		errrrrorkill := Playsound.Process.Kill()
		if errrrrorkill != nil {
			fmt.Println("ERROR")
		}
	}
}
