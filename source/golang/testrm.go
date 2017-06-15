package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var Soundname string = "01 K..mp3"
	rmerror := exec.Command("rm", "public/assets/"+Soundname).Run()
	if rmerror != nil {
		fmt.Println("ERROR rm")
	}
}
