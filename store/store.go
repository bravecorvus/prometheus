// Package store is a small bbolt-backed persistence layer for Prometheus.
//
// All user-configurable state — alarms and individual settings — lives in a
// single bbolt file. Writes are atomic and durable across power loss, which
// matters because Prometheus runs on a Raspberry Pi where SIGKILL or sudden
// power-off is a normal end-of-day event.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"prometheus/structs"
)

const (
	alarmsBucket   = "alarms"
	settingsBucket = "settings"

	// Setting keys.
	KeyEmail           = "email"
	KeyEnableEmail     = "enable_email"
	KeyEnableLed       = "enable_led"
	KeyCustomSoundcard = "custom_soundcard"
	KeyColors          = "colors"
	KeyLastIP          = "last_ip"
)

type Store struct {
	db *bbolt.DB
}

// Open opens (or creates) the bbolt database at path and ensures required
// buckets exist. The caller must Close() the returned store.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	db, err := bbolt.Open(path, 0o644, nil)
	if err != nil {
		return nil, fmt.Errorf("open bbolt: %w", err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, name := range []string{alarmsBucket, settingsBucket} {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("init buckets: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// LoadAlarms returns all stored alarms keyed by name, in name order.
func (s *Store) LoadAlarms() ([]structs.Alarm, error) {
	var alarms []structs.Alarm
	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(alarmsBucket))
		return b.ForEach(func(_, v []byte) error {
			var a structs.Alarm
			if err := json.Unmarshal(v, &a); err != nil {
				return err
			}
			alarms = append(alarms, a)
			return nil
		})
	})
	return alarms, err
}

func (s *Store) SaveAlarm(a structs.Alarm) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(alarmsBucket))
		data, err := json.Marshal(a)
		if err != nil {
			return err
		}
		return b.Put([]byte(a.Name), data)
	})
}

func (s *Store) SaveAlarms(alarms []structs.Alarm) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(alarmsBucket))
		for _, a := range alarms {
			data, err := json.Marshal(a)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(a.Name), data); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetString returns the setting value, or "" if unset.
func (s *Store) GetString(key string) string {
	var out string
	s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(settingsBucket)).Get([]byte(key))
		if v != nil {
			out = string(v)
		}
		return nil
	})
	return out
}

func (s *Store) PutString(key, val string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(settingsBucket)).Put([]byte(key), []byte(val))
	})
}

// GetBool returns true if the setting is stored as "true". Anything else
// (including absent) returns false.
func (s *Store) GetBool(key string) bool {
	return s.GetString(key) == "true"
}

func (s *Store) PutBool(key string, v bool) error {
	if v {
		return s.PutString(key, "true")
	}
	return s.PutString(key, "false")
}

// MigrateLegacyAlarms imports the old alarms.json into the alarms bucket if
// the bucket is empty. On success, the legacy file is renamed with a
// ".migrated" suffix so future starts skip the import. Missing source file is
// not an error.
func (s *Store) MigrateLegacyAlarms(legacyPath string) error {
	empty, err := s.alarmsBucketEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	raw, err := os.ReadFile(legacyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read legacy alarms: %w", err)
	}

	type legacyAlarm struct {
		Name      string `json:"name"`
		Time      string `json:"time"`
		Sound     string `json:"sound"`
		Vibration string `json:"vibration"`
	}
	var legacy []legacyAlarm
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return fmt.Errorf("parse legacy alarms: %w", err)
	}

	alarms := make([]structs.Alarm, len(legacy))
	for i, l := range legacy {
		alarms[i] = structs.Alarm{
			Name:      l.Name,
			Alarmtime: l.Time,
			Sound:     l.Sound == "on",
			Vibration: l.Vibration == "on",
		}
	}
	if err := s.SaveAlarms(alarms); err != nil {
		return err
	}
	return os.Rename(legacyPath, legacyPath+".migrated")
}

// MigrateLegacySetting reads a single-value legacy file and writes its
// trimmed contents to the given setting key, but only if the key is currently
// unset. Missing source file is not an error.
func (s *Store) MigrateLegacySetting(legacyPath, key string) error {
	if s.GetString(key) != "" {
		return nil
	}
	raw, err := os.ReadFile(legacyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read legacy %s: %w", key, err)
	}
	val := string(raw)
	// Strip trailing newline / whitespace; original files were written with
	// no consistent format.
	for len(val) > 0 && (val[len(val)-1] == '\n' || val[len(val)-1] == '\r' || val[len(val)-1] == ' ') {
		val = val[:len(val)-1]
	}
	// Legacy enableemail used "True" (capitalized); normalize.
	if val == "True" {
		val = "true"
	}
	if val == "False" {
		val = "false"
	}
	if err := s.PutString(key, val); err != nil {
		return err
	}
	return os.Rename(legacyPath, legacyPath+".migrated")
}

// SeedDefaultAlarms inserts four placeholder alarms if the bucket is empty.
// Used on fresh installs where there is no legacy alarms.json to migrate.
func (s *Store) SeedDefaultAlarms() error {
	empty, err := s.alarmsBucketEmpty()
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}
	defaults := []structs.Alarm{
		{Name: "alarm1", Alarmtime: "08:00"},
		{Name: "alarm2", Alarmtime: "08:00"},
		{Name: "alarm3", Alarmtime: "08:00"},
		{Name: "alarm4", Alarmtime: "08:00"},
	}
	return s.SaveAlarms(defaults)
}

func (s *Store) alarmsBucketEmpty() (bool, error) {
	empty := true
	err := s.db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(alarmsBucket)).ForEach(func(_, _ []byte) error {
			empty = false
			return nil
		})
	})
	return empty, err
}
