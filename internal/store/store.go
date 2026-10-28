// Package store persists learner state (progress, bookmarks, quiz scores,
// mock attempts, settings) and code drafts as plain files on disk.
package store

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// Store keeps progress.json plus one file per code draft. The state is opaque
// to the server beyond being valid JSON: the UI owns its shape.
type Store struct {
	dir string
	mu  sync.Mutex
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(dir, "drafts"), 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) statePath() string { return filepath.Join(s.dir, "progress.json") }

// LoadState returns the saved state, or "{}" when nothing was saved yet.
func (s *Store) LoadState() (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.statePath())
	if errors.Is(err, os.ErrNotExist) {
		return json.RawMessage("{}"), nil
	}
	if err != nil {
		return nil, err
	}
	if !json.Valid(b) {
		return json.RawMessage("{}"), nil
	}
	return b, nil
}

// SaveState atomically replaces the state file.
func (s *Store) SaveState(b []byte) error {
	if !json.Valid(b) {
		return errors.New("state must be valid JSON")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeAtomic(s.statePath(), b)
}

var safeKey = regexp.MustCompile(`^[a-zA-Z0-9:_\-.]{1,200}$`)

func (s *Store) draftPath(key string) (string, error) {
	if !safeKey.MatchString(key) {
		return "", errors.New("invalid draft key")
	}
	sum := sha1.Sum([]byte(key))
	return filepath.Join(s.dir, "drafts", hex.EncodeToString(sum[:8])+".txt"), nil
}

func (s *Store) LoadDraft(key string) (string, bool, error) {
	p, err := s.draftPath(key)
	if err != nil {
		return "", false, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	return string(b), err == nil, err
}

func (s *Store) SaveDraft(key, code string) error {
	p, err := s.draftPath(key)
	if err != nil {
		return err
	}
	return writeAtomic(p, []byte(code))
}

func (s *Store) DeleteDraft(key string) error {
	p, err := s.draftPath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func writeAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
