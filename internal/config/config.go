package config

import (
	"encoding/json"
	"os"
)

// LoadSettings loads settings.json or returns defaults
func LoadSettings(p Paths) Settings {
	s := DefaultSettings()
	data, err := os.ReadFile(p.SettingsFile())
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	if s.MaxHistory <= 0 {
		s.MaxHistory = 500
	}
	return s
}

// SaveSettings writes settings.json
func SaveSettings(p Paths, s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.SettingsFile(), data, 0644)
}

// LoadEndpoints loads endpoints.json or returns defaults
func LoadEndpoints(p Paths) EndpointsConfig {
	e := DefaultEndpoints()
	data, err := os.ReadFile(p.EndpointsFile())
	if err != nil {
		return e
	}
	_ = json.Unmarshal(data, &e)
	return e
}

// SaveEndpoints writes endpoints.json
func SaveEndpoints(p Paths, e EndpointsConfig) error {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.EndpointsFile(), data, 0644)
}

// LoadHistory loads history.json or returns empty
func LoadHistory(p Paths) History {
	h := History{}
	data, err := os.ReadFile(p.HistoryFile())
	if err != nil {
		return h
	}
	_ = json.Unmarshal(data, &h)
	return h
}

// SaveHistory writes history.json
func SaveHistory(p Paths, h History) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.HistoryFile(), data, 0644)
}

// AddHistoryEntry adds an entry to the LRU history, deduplicating
func AddHistoryEntry(h *History, entry string, max int) {
	// Remove duplicate if exists
	entries := make([]string, 0, len(h.Entries))
	for _, e := range h.Entries {
		if e != entry {
			entries = append(entries, e)
		}
	}
	entries = append(entries, entry)
	// Trim from front if over max
	if len(entries) > max {
		entries = entries[len(entries)-max:]
	}
	h.Entries = entries
}
