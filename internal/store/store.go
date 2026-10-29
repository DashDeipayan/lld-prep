// Package store persists learner state (progress, bookmarks, quiz scores,
// mock attempts, settings) and code drafts. The state is opaque to the server
// beyond being a JSON object: the UI owns its shape.
package store

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// Store is implemented by Files (single learner, plain files on disk) and by
// accounts.Mongo (one record per signed-in user).
type Store interface {
	LoadState(ctx context.Context, user string) (json.RawMessage, error)
	SaveState(ctx context.Context, user string, b []byte) error
	LoadDraft(ctx context.Context, user, key string) (string, bool, error)
	SaveDraft(ctx context.Context, user, key, code string) error
	DeleteDraft(ctx context.Context, user, key string) error
}

var (
	ErrInvalidState = errors.New("state must be a JSON object")
	ErrInvalidKey   = errors.New("invalid draft key")
)

var safeKey = regexp.MustCompile(`^[a-zA-Z0-9:_\-.]{1,200}$`)

// ValidKey reports whether a draft key is safe to use as a file name or id.
func ValidKey(key string) bool { return safeKey.MatchString(key) }

// ValidState reports whether b is a JSON object.
func ValidState(b []byte) bool {
	var m map[string]json.RawMessage
	return json.Unmarshal(b, &m) == nil && m != nil
}

// Files keeps progress.json plus one file per code draft. It serves a single
// learner, so the user argument is ignored.
type Files struct {
	dir string
	mu  sync.Mutex
}

func OpenFiles(dir string) (*Files, error) {
	if err := os.MkdirAll(filepath.Join(dir, "drafts"), 0o755); err != nil {
		return nil, err
	}
	return &Files{dir: dir}, nil
}

func (s *Files) statePath() string { return filepath.Join(s.dir, "progress.json") }

// LoadState returns the saved state, or "{}" when nothing was saved yet.
func (s *Files) LoadState(_ context.Context, _ string) (json.RawMessage, error) {
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
func (s *Files) SaveState(_ context.Context, _ string, b []byte) error {
	if !ValidState(b) {
		return ErrInvalidState
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return writeAtomic(s.statePath(), b)
}

func (s *Files) draftPath(key string) (string, error) {
	if !ValidKey(key) {
		return "", ErrInvalidKey
	}
	sum := sha1.Sum([]byte(key))
	return filepath.Join(s.dir, "drafts", hex.EncodeToString(sum[:8])+".txt"), nil
}

func (s *Files) LoadDraft(_ context.Context, _, key string) (string, bool, error) {
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

func (s *Files) SaveDraft(_ context.Context, _, key, code string) error {
	p, err := s.draftPath(key)
	if err != nil {
		return err
	}
	return writeAtomic(p, []byte(code))
}

func (s *Files) DeleteDraft(_ context.Context, _, key string) error {
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
