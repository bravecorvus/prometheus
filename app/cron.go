package app

import (
	"fmt"
	"os/exec"
	"prometheus/gpio"
	"prometheus/nixie"
	"prometheus/structs"
	"prometheus/utils"
	"time"
)

func (app *App) SendTime() {
	timeStr := nixie.CurrentTimeAsString()

	if app.FoundNixie {
		var payload string
		if app.EnableLed {
			payload = timeStr + app.Red + app.Green + app.Blue
		} else {
			payload = timeStr
		}
		if _, err := app.Port.Write([]byte(payload)); err != nil {
			fmt.Println(err.Error())
		}
	} else {
		app.Options.PortName = nixie.FindArduino()
		app.FoundNixie = app.Options.PortName != ""
	}
}

func (app *App) AlarmLoop() {
	t := time.Now()
	currenttime := t.Format("15:04")

	if app.EnableEmail {
		utils.CheckIPChange()
	}

	for i := range app.Alarms {
		if app.Alarms[i].Alarmtime == currenttime {
			go utils.RestartNetwork()
			app.runAlarm(&app.Alarms[i])
			return
		}
	}
}

// runAlarm handles a single alarm trigger with the appropriate sound/vibration combination.
func (app *App) runAlarm(alarm *structs.Alarm) {
	alarm.CurrentlyRunning = true
	app.setAlarmLED()

	switch {
	case alarm.Sound && alarm.Vibration:
		app.runSoundAndVibration(alarm)
	case alarm.Sound:
		app.runSoundOnly(alarm)
	case alarm.Vibration:
		app.runVibrationOnly(alarm)
	default:
		alarm.CurrentlyRunning = false
	}
}

func (app *App) setAlarmLED() {
	app.Red = "255"
	app.Green = "000"
	app.Blue = "000"
}

func (app *App) resetLED() {
	app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
}

func (app *App) buildPlayCommand() *exec.Cmd {
	if app.CustomSoundCard {
		return exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
	}
	return exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
}

func (app *App) runSoundAndVibration(alarm *structs.Alarm) {
	playsound := app.buildPlayCommand()
	if err := playsound.Start(); err != nil {
		fmt.Println(err.Error())
	}

	duration := time.Second * 3
	for {
		gpio.VibOn()
		stopped := waitForStop(alarm, 50, 50*time.Millisecond)
		if stopped {
			gpio.VibOff()
			killProcess(playsound)
			app.resetLED()
			return
		}
		if utils.OverTenMinutes(alarm.Alarmtime) {
			alarm.CurrentlyRunning = false
			gpio.VibOff()
			killProcess(playsound)
			app.resetLED()
			return
		}
		gpio.VibOff()
		time.Sleep(duration)
	}
}

func (app *App) runSoundOnly(alarm *structs.Alarm) {
	playsound := app.buildPlayCommand()
	if err := playsound.Start(); err != nil {
		fmt.Println(err.Error())
	}

	for {
		time.Sleep(time.Second)
		if !alarm.CurrentlyRunning {
			killProcess(playsound)
			app.resetLED()
			return
		}
		if utils.OverTenMinutes(alarm.Alarmtime) {
			alarm.CurrentlyRunning = false
			killProcess(playsound)
			app.resetLED()
			return
		}
	}
}

func (app *App) runVibrationOnly(alarm *structs.Alarm) {
	duration := time.Second * 3
	for {
		gpio.VibOn()
		stopped := waitForStop(alarm, 50, 50*time.Millisecond)
		if stopped {
			gpio.VibOff()
			app.resetLED()
			return
		}
		if utils.OverTenMinutes(alarm.Alarmtime) {
			alarm.CurrentlyRunning = false
			gpio.VibOff()
			app.resetLED()
			return
		}
		gpio.VibOff()
		time.Sleep(duration)
	}
}

// waitForStop polls alarm.CurrentlyRunning in small increments.
// Returns true if the alarm was stopped by the user.
func waitForStop(alarm *structs.Alarm, iterations int, interval time.Duration) bool {
	for i := 0; i < iterations; i++ {
		time.Sleep(interval)
		if !alarm.CurrentlyRunning {
			return true
		}
	}
	return false
}

func killProcess(cmd *exec.Cmd) {
	if cmd.Process != nil {
		if err := cmd.Process.Kill(); err != nil {
			fmt.Println(err.Error())
		}
	}
}
