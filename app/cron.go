package app

import (
	"fmt"
	"os/exec"
	"prometheus/config"
	"prometheus/gpio"
	"prometheus/nixie"
	"prometheus/structs"
	"prometheus/utils"
	"time"
)

func vibOn() {
	if config.DemoMode {
		DemoHub.Broadcast("vibration_on")
		return
	}
	gpio.VibOn()
}

func vibOff() {
	if config.DemoMode {
		DemoHub.Broadcast("vibration_off")
		return
	}
	gpio.VibOff()
}

// startSound either spawns cvlc (prod) or broadcasts a demo event.
// Returns a stop function the caller invokes when the alarm ends.
func (app *App) startSound() func() {
	if config.DemoMode {
		DemoHub.Broadcast("sound_start")
		return func() { DemoHub.Broadcast("sound_stop") }
	}
	cmd := app.buildPlayCommand()
	if err := cmd.Start(); err != nil {
		fmt.Println(err.Error())
	}
	return func() { killProcess(cmd) }
}

func (app *App) SendTime() {
	if config.DemoMode {
		return
	}
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

	if app.EnableEmail && !config.DemoMode {
		utils.CheckIPChange()
	}

	for i := range app.Alarms {
		if app.Alarms[i].Alarmtime == currenttime {
			if !config.DemoMode {
				go utils.RestartNetwork()
			}
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
	stopSound := app.startSound()

	duration := time.Second * 3
	for {
		vibOn()
		stopped := waitForStop(alarm, 50, 50*time.Millisecond)
		if stopped {
			vibOff()
			stopSound()
			app.resetLED()
			return
		}
		if utils.OverTenMinutes(alarm.Alarmtime) {
			alarm.CurrentlyRunning = false
			vibOff()
			stopSound()
			app.resetLED()
			return
		}
		vibOff()
		time.Sleep(duration)
	}
}

func (app *App) runSoundOnly(alarm *structs.Alarm) {
	stopSound := app.startSound()

	for {
		time.Sleep(time.Second)
		if !alarm.CurrentlyRunning {
			stopSound()
			app.resetLED()
			return
		}
		if utils.OverTenMinutes(alarm.Alarmtime) {
			alarm.CurrentlyRunning = false
			stopSound()
			app.resetLED()
			return
		}
	}
}

func (app *App) runVibrationOnly(alarm *structs.Alarm) {
	duration := time.Second * 3
	for {
		vibOn()
		stopped := waitForStop(alarm, 50, 50*time.Millisecond)
		if stopped {
			vibOff()
			app.resetLED()
			return
		}
		if utils.OverTenMinutes(alarm.Alarmtime) {
			alarm.CurrentlyRunning = false
			vibOff()
			app.resetLED()
			return
		}
		vibOff()
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
