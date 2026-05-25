package app

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"prometheus/config"
	"prometheus/nixie"
	"prometheus/store"
	"prometheus/structs"
	"prometheus/utils"
	"strings"
	"sync"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

const NumAlarms = 4

type App struct {
	Alarms           [NumAlarms]structs.Alarm
	alarmsMu         sync.Mutex
	EnableEmail      bool
	Email            string
	Soundname        string
	CustomSoundCard  bool
	EnableLed        bool
	Red, Green, Blue string
	Options          serial.OpenOptions
	Port             io.ReadWriteCloser
	FoundNixie       bool
	Store            *store.Store
	AssetsDir        string
}

func (app *App) Initialize(s *store.Store, assetsDir string) {
	app.Store = s
	app.AssetsDir = assetsDir

	app.loadAlarmsFromStore()
	app.Soundname = firstAudioFile(assetsDir)

	app.Email = s.GetString(store.KeyEmail)
	app.EnableEmail = s.GetBool(store.KeyEnableEmail)
	app.CustomSoundCard = s.GetBool(store.KeyCustomSoundcard)
	app.Red, app.Green, app.Blue = utils.ParseHexToRGB(s.GetString(store.KeyColors))
	app.EnableLed = s.GetBool(store.KeyEnableLed)

	if config.DemoMode {
		app.FoundNixie = false
		return
	}

	app.Options.PortName = nixie.FindArduino()
	app.Options.BaudRate = 115200
	app.Options.DataBits = 8
	app.Options.StopBits = 1
	app.Options.MinimumReadSize = 4

	time.Sleep(20 * time.Second)

	port, err := serial.Open(app.Options)
	if err != nil {
		app.FoundNixie = false
	} else {
		app.FoundNixie = true
		app.Port = port
	}
}

// loadAlarmsFromStore reads all alarms from bbolt into the fixed-size array.
// bbolt's ForEach returns keys in lexicographical order, which for the
// "alarm1".."alarm4" naming scheme gives natural ordering.
func (app *App) loadAlarmsFromStore() {
	loaded, err := app.Store.LoadAlarms()
	if err != nil {
		fmt.Println("load alarms:", err)
	}
	for i := 0; i < NumAlarms && i < len(loaded); i++ {
		app.Alarms[i] = loaded[i]
		app.Alarms[i].CurrentlyRunning = false
	}
}

// findAlarm returns a pointer to the alarm matching the given name, or nil.
// Caller must hold alarmsMu.
func (app *App) findAlarm(name string) *structs.Alarm {
	for i := range app.Alarms {
		if app.Alarms[i].Name == name {
			return &app.Alarms[i]
		}
	}
	return nil
}

func (app *App) persistAlarms() {
	if err := app.Store.SaveAlarms(app.Alarms[:]); err != nil {
		fmt.Println("persist alarms:", err)
	}
}

// firstAudioFile returns the first filename in dir, or "" if dir is empty or
// unreadable. The frontend treats "" as "no track uploaded".
func firstAudioFile(dir string) string {
	var b bytes.Buffer
	if err := utils.Execute(&b, exec.Command("ls", dir)); err != nil {
		return ""
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return ""
	}
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		out = out[:i]
	}
	return out
}
