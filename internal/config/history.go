package config

import (
	"encoding/json"
	"os"
)

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
