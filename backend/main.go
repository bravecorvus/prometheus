// prometheus is the alarm-clock backend. It exposes a small JSON API
// consumed by the React frontend (served from ./dist) and runs two cron
// loops: one ticks the Nixie display every second, one checks alarms every
// minute. Persistent state lives in ./data/prometheus.db (bbolt).
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"prometheus/app"
	"prometheus/config"
	"prometheus/store"

	"github.com/robfig/cron"
)

// Layout under the working directory. On the Pi this is set to
// /home/pi/prometheus by the systemd unit, matching the repo layout exactly:
// backend/prometheus is the binary, frontend/dist holds the built UI, and
// data/ holds runtime state shared between dev and prod.
const (
	distDir   = "frontend/dist"
	assetsDir = "data/assets"
	dbFile    = "data/prometheus.db"
)

var globals app.App

func main() {
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		fmt.Println("create assets dir:", err)
		os.Exit(1)
	}

	s, err := store.Open(dbFile)
	if err != nil {
		fmt.Println("open store:", err)
		os.Exit(1)
	}
	defer s.Close()

	if err := s.SeedDefaultAlarms(); err != nil {
		fmt.Println("seed alarms:", err)
	}

	globals.Initialize(s, assetsDir)

	c := cron.New()
	if !config.DemoMode {
		c.AddFunc("@every 1s", func() { globals.SendTime() })
	}
	c.AddFunc("0 * * * * *", func() { globals.AlarmLoop() })
	c.Start()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/state", globals.StateHandler)
	mux.HandleFunc("PATCH /api/alarms/{name}", globals.AlarmPatchHandler)
	mux.HandleFunc("POST /api/snooze", globals.SnoozeHandler)
	mux.HandleFunc("PATCH /api/settings", globals.SettingsPatchHandler)
	mux.HandleFunc("POST /api/upload", globals.UploadHandler)
	mux.HandleFunc("GET /ws", globals.ServeWS)
	mux.Handle("GET /audio/", http.StripPrefix("/audio/", http.FileServer(http.Dir(assetsDir))))
	mux.HandleFunc("/", spaHandler)

	srv := &http.Server{Addr: ":3000", Handler: mux}
	go func() {
		fmt.Println("listening on :3000")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("http server:", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// spaHandler serves the built frontend with SPA fallback: anything that
// doesn't match a real file in dist/ returns index.html so client-side routes
// resolve. If dist/ is missing (running the backend bare during dev), it
// returns a helpful 503.
func spaHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(distDir); err != nil {
		http.Error(w, "frontend not built: run `pnpm --dir frontend build` (or `pnpm --dir frontend dev` for hot reload on :5173)", http.StatusServiceUnavailable)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/ws" {
		http.NotFound(w, r)
		return
	}
	clean := filepath.Clean(r.URL.Path)
	if clean == "/" {
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		return
	}
	full := filepath.Join(distDir, clean)
	if _, err := os.Stat(full); err != nil {
		http.ServeFile(w, r, filepath.Join(distDir, "index.html"))
		return
	}
	http.ServeFile(w, r, full)
}
