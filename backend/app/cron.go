package app

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"prometheus/config"
	"prometheus/gpio"
	"prometheus/nixie"
	"prometheus/store"
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
	if !app.FoundNixie {
		app.Options.PortName = nixie.FindArduino()
		app.FoundNixie = app.Options.PortName != ""
		return
	}

	// Always send a fixed 15-byte payload: HHMMSSRRRGGGBBB. The Arduino sketch
	// has no framing (no newline, no length prefix), so a variable-length send
	// desyncs the reader — one 6-byte "LED off" tick leaves 9 stale bytes in
	// the buffer and the next tick's HHMMSS gets interpreted as color values,
	// which is why LEDs light up randomly. When the user disables LEDs we
	// still send RGB, just with "000000000".
	timeStr := nixie.CurrentTimeAsString()
	r, g, b := "000", "000", "000"
	if app.EnableLed {
		r, g, b = app.Red, app.Green, app.Blue
	}
	payload := timeStr + r + g + b

	app.portMu.Lock()
	_, err := app.Port.Write([]byte(payload))
	app.portMu.Unlock()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (app *App) AlarmLoop() {
	t := time.Now()
	currenttime := t.Format("15:04")

	if app.EnableEmail && !config.DemoMode {
		app.checkIPChange()
	}

	for i := range app.Alarms {
		if app.Alarms[i].Alarmtime == currenttime {
			app.runAlarm(&app.Alarms[i])
			return
		}
	}
}

// checkIPChange compares the current wlan0 IP against the last-known IP in
// the store. If it changed, persist the new value and email it out.
func (app *App) checkIPChange() {
	current := utils.GetIP()
	if last := app.Store.GetString(store.KeyLastIP); last != current {
		if err := app.Store.PutString(store.KeyLastIP, current); err != nil {
			fmt.Println("save last_ip:", err)
		}
		send := exec.Command("backend/email/prometheusemail", app.Email, current)
		if err := send.Run(); err != nil {
			fmt.Println("failed to send email")
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
	app.Red, app.Green, app.Blue = utils.ParseHexToRGB(app.Store.GetString(store.KeyColors))
	app.EnableLed = app.Store.GetBool(store.KeyEnableLed)
}

func (app *App) buildPlayCommand() *exec.Cmd {
	track := filepath.Join(app.AssetsDir, app.Soundname)
	if app.CustomSoundCard {
		return exec.Command("cvlc", track, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
	}
	return exec.Command("cvlc", track, "--gain=0.04")
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
