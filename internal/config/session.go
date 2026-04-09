package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewSessionFile creates a new empty session
func NewSessionFile(provider, model string, thinking bool, systemPrompt string) SessionFile {
	now := time.Now().UTC().Format(time.RFC3339)
	return SessionFile{
		Version: 1,
		Session: Session{
			ID:               uuid.New().String(),
			CreatedAt:        now,
			UpdatedAt:        now,
			Provider:         provider,
			Model:            model,
			Thinking:         thinking,
			SystemPromptFile: systemPrompt,
		},
	}
}

// SaveSession writes a session to sessions/<name>.chat.json
func SaveSession(p Paths, name string, sf SessionFile) error {
	sf.Session.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(sf, "", "  ")
	if err != nil {
		return err
	}
	file := filepath.Join(p.Sessions, name+".chat.json")
	return os.WriteFile(file, data, 0644)
}

// LoadSession reads a session from sessions/<name>.chat.json
func LoadSession(p Paths, name string) (SessionFile, error) {
	file := filepath.Join(p.Sessions, name+".chat.json")
	data, err := os.ReadFile(file)
	if err != nil {
		return SessionFile{}, err
	}
	var sf SessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return SessionFile{}, err
	}
	return sf, nil
}

// ListSessions returns available session names (without .chat.json)
func ListSessions(p Paths) []string {
	entries, err := os.ReadDir(p.Sessions)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".chat.json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".chat.json"))
		}
	}
	return names
}
