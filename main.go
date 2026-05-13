// prometheus is the main program logic that runs both the alarm logic and the web server hosting the front-end user interface for users to control the Prometheus. The main function is split into two parts: First is the part that runs two cron jobs: the first runs once every second to try to send the time + LED string to the nixie clock and the second once every minute to check if the current time matches a user supplied alarm (also running the releavnt alarm actions vibrate and or output sound based on the user set parameters). Then the second half deals with providing the web server functionality. The fileserver serves a plain HTML file whose client side scripting uses Vue.js, Bootstrap, and jQuery. Persistent state (alarms, settings) lives in a single bbolt database under ./data/prometheus.db; the frontend still fetches /json/<name> URLs, but those are now served by handlers reading from the store rather than from static files on disk.

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prometheus/app"
	"prometheus/config"
	"prometheus/store"
	"prometheus/utils"

	"github.com/robfig/cron"
)

var globals app.App

func main() {
	jsonDir := utils.Pwd() + "/public/json"
	dbPath := utils.Pwd() + "/data/prometheus.db"

	s, err := store.Open(dbPath)
	if err != nil {
		fmt.Println("failed to open store:", err)
		os.Exit(1)
	}
	defer s.Close()

	migrateLegacyState(s, jsonDir)

	// SeedDefaultAlarms only runs if the bucket is still empty after migration.
	if err := s.SeedDefaultAlarms(); err != nil {
		fmt.Println("seed default alarms:", err)
	}

	globals.Initialize(s)

	c := cron.New()
	if !config.DemoMode {
		c.AddFunc("@every 1s", func() { globals.SendTime() })
	}
	c.AddFunc("0 * * * * *", func() { globals.AlarmLoop() })
	c.Start()

	// Read endpoints the frontend polls on page load. These must be registered
	// before the catch-all file server so the mux dispatches them first.
	http.HandleFunc("/json/alarms.json", globals.AlarmsReadHandler)
	http.HandleFunc("/json/email", globals.EmailReadHandler)
	http.HandleFunc("/json/enableemail", globals.EnableEmailReadHandler)
	http.HandleFunc("/json/colors", globals.ColorsReadHandler)
	http.HandleFunc("/json/customsoundcard", globals.CustomSoundcardReadHandler)
	http.HandleFunc("/json/enableled", globals.EnableLedReadHandler)

	// Write endpoints.
	http.HandleFunc("/time", globals.TimeHandler)
	http.HandleFunc("/sound", globals.SoundHandler)
	http.HandleFunc("/vibration", globals.VibrationHandler)
	http.HandleFunc("/snooze", globals.SnoozeHandler)
	http.HandleFunc("/enableemail", globals.EnableEmailHandler)
	http.HandleFunc("/customsoundcard", globals.CustomSoundcardHandler)
	http.HandleFunc("/newemail", globals.NewEmailHandler)
	http.HandleFunc("/submitcolors", globals.SubmitColorsHandler)
	http.HandleFunc("/submitenableled", globals.SubmitEnableLEDHandler)
	http.HandleFunc("/upload", globals.UploadHandler)
	http.HandleFunc("/api/mode", globals.ModeHandler)
	http.HandleFunc("/ws", globals.ServeWS)

	// Everything else falls through to static files.
	fs := http.FileServer(http.Dir(utils.Pwd() + "/public"))
	http.Handle("/", fs)

	srv := &http.Server{Addr: ":3000"}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("http server:", err)
		}
	}()

	// Graceful shutdown so the store is closed on SIGINT/SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// migrateLegacyState reads the old per-file state into bbolt on first run. Any
// successfully imported file is renamed to <file>.migrated so the import is
// idempotent and the originals stick around for rollback.
func migrateLegacyState(s *store.Store, jsonDir string) {
	if err := s.MigrateLegacyAlarms(jsonDir + "/alarms.json"); err != nil {
		fmt.Println("migrate alarms:", err)
	}
	migrations := []struct {
		path string
		key  string
	}{
		{jsonDir + "/email", store.KeyEmail},
		{jsonDir + "/enableemail", store.KeyEnableEmail},
		{jsonDir + "/customsoundcard", store.KeyCustomSoundcard},
		{jsonDir + "/enableled", store.KeyEnableLed},
		{jsonDir + "/colors", store.KeyColors},
		{jsonDir + "/ip", store.KeyLastIP},
	}
	for _, m := range migrations {
		if err := s.MigrateLegacySetting(m.path, m.key); err != nil {
			fmt.Println("migrate", m.key+":", err)
		}
	}
}
