package app

import (
	"fmt"
	"os/exec"
	"prometheus/gpio"
	"prometheus/nixie"
	"prometheus/utils"
	"time"
)

func (app *App) SendTime() {

	// fmt.Println("RGB(", Red, Green, Blue, ")")

	if app.EnableLed {
		if app.FoundNixie {
			b := []byte(nixie.CurrentTimeAsString() + app.Red + app.Green + app.Blue)
			_, err := app.Port.Write(b)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			app.Options.PortName = nixie.FindArduino()
			if app.Options.PortName != "" {
				app.FoundNixie = true
			} else {
				app.FoundNixie = false
			}
		}

	} else {

		if app.FoundNixie {
			b := []byte(nixie.CurrentTimeAsString())
			_, err := app.Port.Write(b)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			app.Options.PortName = nixie.FindArduino()
			if app.Options.PortName != "" {
				app.FoundNixie = true
			} else {
				app.FoundNixie = false
			}
		}
	}

}

func (app *App) AlarmLoop() {
	breaktime := false
	duration := time.Second * 3
	t := time.Now()
	currenttime := t.Format("15:04")
	if app.EnableEmail {
		utils.CheckIPChange()
	}

	if app.Alarm1.Alarmtime == currenttime {

		go utils.RestartNetwork()
		app.Alarm1.CurrentlyRunning = true

		if app.Alarm1.Sound && app.Alarm1.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")

				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm1.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {

						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm1.Alarmtime) {
						app.Alarm1.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}
			} else {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}
				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm1.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {

						gpio.VibOff()
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println(errrrrorkill.Error())
						}
						breaktime = false
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm1.Alarmtime) {
						app.Alarm1.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}
			}

		} else if app.Alarm1.Sound && !app.Alarm1.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"
			if app.CustomSoundCard {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}
				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm1.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm1.Alarmtime) {
						app.Alarm1.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					}
				}
			} else {

				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}
				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm1.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm1.Alarmtime) {
						app.Alarm1.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}
			}

		} else if !app.Alarm1.Sound && app.Alarm1.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			for {
				gpio.VibOn()
				for i := 1; i <= 50; i++ {
					time.Sleep(time.Millisecond * 50)
					if !app.Alarm1.CurrentlyRunning {
						breaktime = true
						break
					}
				}
				if breaktime {
					gpio.VibOff()
					breaktime = false
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break

				} else if utils.OverTenMinutes(app.Alarm1.Alarmtime) {
					app.Alarm1.CurrentlyRunning = false
					gpio.VibOff()
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break
				} else {
					gpio.VibOff()
					time.Sleep(duration)
				}
			}
		} else {
			app.Alarm1.CurrentlyRunning = false
		}

	} else if app.Alarm2.Alarmtime == currenttime {

		// Check if there is network connectivity (if not, then restart network interfaces)
		go utils.RestartNetwork()
		app.Alarm2.CurrentlyRunning = true

		if app.Alarm2.Sound && app.Alarm2.Vibration {
			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {

				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm2.CurrentlyRunning {
							breaktime = true
							app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm2.Alarmtime) {
						app.Alarm2.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}
			} else {

				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}
				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm2.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm2.Alarmtime) {
						app.Alarm2.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}
			}

		} else if app.Alarm2.Sound && !app.Alarm2.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {

				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm2.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm2.Alarmtime) {
						app.Alarm2.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}

			} else {

				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}
				for {

					time.Sleep(time.Second * 1)
					if !app.Alarm2.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm2.Alarmtime) {
						app.Alarm2.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					}
				}
			}

		} else if !app.Alarm2.Sound && app.Alarm2.Vibration {
			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			for {
				gpio.VibOn()
				for i := 1; i <= 50; i++ {
					time.Sleep(time.Millisecond * 50)
					if !app.Alarm2.CurrentlyRunning {
						breaktime = true
						break
					}
				}
				if breaktime {
					gpio.VibOff()
					breaktime = false
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break

				} else if utils.OverTenMinutes(app.Alarm2.Alarmtime) {
					app.Alarm2.CurrentlyRunning = false
					gpio.VibOff()
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break
				} else {
					gpio.VibOff()
					time.Sleep(duration)
				}
			}
		} else {
			app.Alarm2.CurrentlyRunning = false
		}

	} else if app.Alarm3.Alarmtime == currenttime {
		// Check if there is network connectivity (if not, then restart network interfaces)
		go utils.RestartNetwork()
		app.Alarm3.CurrentlyRunning = true

		if app.Alarm3.Sound && app.Alarm3.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm3.Alarmtime) {
						app.Alarm3.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}
			} else {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm3.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm3.Alarmtime) {
						app.Alarm3.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}
			}

		} else if app.Alarm3.Sound && !app.Alarm3.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm3.CurrentlyRunning {
						errrrrorkill := playsound.Process.Kill()
						if errrrrorkill != nil {
							fmt.Println(errrrrorkill.Error())
						}
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()

						break
					} else if utils.OverTenMinutes(app.Alarm3.Alarmtime) {
						app.Alarm3.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}

			} else {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm3.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm3.Alarmtime) {
						app.Alarm3.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}

			}

		} else if !app.Alarm3.Sound && app.Alarm3.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			for {
				gpio.VibOn()
				for i := 1; i <= 50; i++ {
					time.Sleep(time.Millisecond * 50)
					if !app.Alarm3.CurrentlyRunning {
						breaktime = true
						break
					}
				}
				if breaktime {
					gpio.VibOff()
					breaktime = false
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break
				} else if utils.OverTenMinutes(app.Alarm3.Alarmtime) {
					app.Alarm3.CurrentlyRunning = false
					gpio.VibOff()
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break
				} else {
					gpio.VibOff()
					time.Sleep(duration)
				}
			}
		} else {
			app.Alarm3.CurrentlyRunning = false
		}

	} else if app.Alarm4.Alarmtime == currenttime {
		// Check if there is network connectivity (if not, then restart network interfaces)
		go utils.RestartNetwork()
		app.Alarm4.CurrentlyRunning = true

		if app.Alarm4.Sound && app.Alarm4.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm4.Alarmtime) {
						app.Alarm4.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}

			} else {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					gpio.VibOn()
					for i := 1; i <= 50; i++ {
						time.Sleep(time.Millisecond * 50)
						if !app.Alarm4.CurrentlyRunning {
							breaktime = true
							break
						}
					}
					if breaktime {
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}
						breaktime = false

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm4.Alarmtime) {
						app.Alarm4.CurrentlyRunning = false
						gpio.VibOff()
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else {
						gpio.VibOff()
						time.Sleep(duration)
					}

				}

			}

		} else if app.Alarm4.Sound && !app.Alarm4.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			if app.CustomSoundCard {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04", "-A=alsa", "--alsa-audio-device=default")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm4.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break

					} else if utils.OverTenMinutes(app.Alarm4.Alarmtime) {
						app.Alarm4.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}

			} else {
				var playsound = exec.Command("cvlc", utils.Pwd()+"/public/assets/"+app.Soundname, "--gain=0.04")
				if err := playsound.Start(); err != nil {
					fmt.Println(err.Error())
				}

				for {
					time.Sleep(time.Second * 1)
					if !app.Alarm4.CurrentlyRunning {
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					} else if utils.OverTenMinutes(app.Alarm4.Alarmtime) {
						app.Alarm4.CurrentlyRunning = false
						if err := playsound.Process.Kill(); err != nil {
							fmt.Println(err.Error())
						}

						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}

			}

		} else if !app.Alarm4.Sound && app.Alarm4.Vibration {

			app.Red = "255"
			app.Green = "000"
			app.Blue = "000"

			for {
				gpio.VibOn()
				for i := 1; i <= 50; i++ {
					time.Sleep(time.Millisecond * 50)
					if !app.Alarm4.CurrentlyRunning {
						breaktime = true
						app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
						break
					}
				}
				if breaktime {
					gpio.VibOff()
					breaktime = false
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break
				} else if utils.OverTenMinutes(app.Alarm4.Alarmtime) {
					app.Alarm4.CurrentlyRunning = false
					gpio.VibOff()
					app.Red, app.Green, app.Blue, app.EnableLed = utils.ColorInitialize()
					break
				} else {
					gpio.VibOff()
					time.Sleep(duration)
				}
			}
		} else {
			app.Alarm4.CurrentlyRunning = false
		}
	}
}
