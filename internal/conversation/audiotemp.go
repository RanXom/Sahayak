package conversation

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AudioStore persists ephemeral TTS response audio. User input audio is never
// stored; only synthesized responses are held briefly for playback.
type AudioStore interface {
	Save(data []byte, format string) (string, error)
	Path(id string) (string, bool)
}

// FileAudioStore stores audio under a directory.
type FileAudioStore struct {
	dir string
}

// NewFileAudioStore creates the directory and returns a store.
func NewFileAudioStore(dir string) (*FileAudioStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &FileAudioStore{dir: dir}, nil
}

// Save writes audio and returns its identifier (without extension).
func (s *FileAudioStore) Save(data []byte, format string) (string, error) {
	format = strings.TrimPrefix(format, ".")
	if format == "" {
		format = "wav"
	}
	id := uuid.NewString()
	name := id + "." + format
	if err := os.WriteFile(filepath.Join(s.dir, name), data, 0o644); err != nil {
		return "", err
	}
	return name, nil
}

// Path resolves an id to a file path, guarding against traversal.
func (s *FileAudioStore) Path(id string) (string, bool) {
	if id == "" || strings.ContainsAny(id, "/\\") {
		return "", false
	}
	full := filepath.Join(s.dir, id)
	if _, err := os.Stat(full); err != nil {
		return "", false
	}
	return full, true
}

// Cleanup removes audio files older than maxAge.
func (s *FileAudioStore) Cleanup(maxAge time.Duration) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(s.dir, e.Name()))
	}
}
