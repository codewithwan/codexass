package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"codexass/internal/common"
	"codexass/internal/config"
	"codexass/internal/domain"
)

type Store struct {
	baseDir    string
	pendingDir string
	statePath  string
}

func New(dir string) (*Store, error) {
	if dir == "" {
		dir = config.DefaultStoreDir()
	}
	store := &Store{
		baseDir:    dir,
		pendingDir: filepath.Join(dir, "pending-oauth"),
		statePath:  filepath.Join(dir, "auth.json"),
	}
	if err := os.MkdirAll(store.pendingDir, 0o700); err != nil {
		return nil, err
	}
	common.MaybeMakePrivate(store.baseDir, 0o700)
	common.MaybeMakePrivate(store.pendingDir, 0o700)
	if _, err := os.Stat(store.statePath); os.IsNotExist(err) {
		if err := store.SaveState(domain.StateFile{Version: 1, Sessions: []domain.SessionRecord{}}); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *Store) LoadState() (domain.StateFile, error) {
	raw, err := os.ReadFile(s.statePath)
	if err != nil {
		return domain.StateFile{}, err
	}
	if len(raw) == 0 {
		return domain.StateFile{Version: 1, Sessions: []domain.SessionRecord{}}, nil
	}
	var state domain.StateFile
	if err := json.Unmarshal(raw, &state); err != nil {
		return domain.StateFile{}, err
	}
	if state.Version == 0 {
		state.Version = 1
	}
	sort.Slice(state.Sessions, func(i, j int) bool {
		return state.Sessions[i].UpdatedAt > state.Sessions[j].UpdatedAt
	})
	return state, nil
}

func (s *Store) SaveState(state domain.StateFile) error {
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.statePath, raw, 0o600); err != nil {
		return err
	}
	common.MaybeMakePrivate(s.statePath, 0o600)
	return nil
}

func (s *Store) SavePending(pending domain.PendingLogin) error {
	path := filepath.Join(s.pendingDir, pending.LoginID+".json")
	raw, err := json.MarshalIndent(pending, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return err
	}
	common.MaybeMakePrivate(path, 0o600)
	return nil
}

func (s *Store) LoadPending(loginID string) (domain.PendingLogin, error) {
	raw, err := os.ReadFile(filepath.Join(s.pendingDir, loginID+".json"))
	if err != nil {
		return domain.PendingLogin{}, err
	}
	var pending domain.PendingLogin
	if err := json.Unmarshal(raw, &pending); err != nil {
		return domain.PendingLogin{}, err
	}
	return pending, nil
}

func (s *Store) DeletePending(loginID string) {
	_ = os.Remove(filepath.Join(s.pendingDir, loginID+".json"))
}

func (s *Store) UpsertSession(session domain.SessionRecord, makeActive bool) (domain.SessionRecord, error) {
	state, err := s.LoadState()
	if err != nil {
		return domain.SessionRecord{}, err
	}
	now := common.ISOFormat(common.UTCNow())
	session.UpdatedAt = now
	if session.CreatedAt == "" {
		session.CreatedAt = now
	}
	items := make([]domain.SessionRecord, 0, len(state.Sessions)+1)
	for _, item := range state.Sessions {
		if item.ID != session.ID {
			items = append(items, item)
		}
	}
	items = append(items, session)
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt > items[j].UpdatedAt })
	state.Sessions = items
	if makeActive || state.ActiveSessionID == "" {
		state.ActiveSessionID = session.ID
	}
	if err := s.SaveState(state); err != nil {
		return domain.SessionRecord{}, err
	}
	return s.FindSession(session.ID)
}

func (s *Store) FindSession(selector string) (domain.SessionRecord, error) {
	state, err := s.LoadState()
	if err != nil {
		return domain.SessionRecord{}, err
	}
	target := strings.TrimSpace(selector)
	if target == "" {
		target = state.ActiveSessionID
	}
	if target == "" {
		return domain.SessionRecord{}, fmt.Errorf("no active session configured")
	}
	needle := strings.ToLower(target)
	for _, item := range state.Sessions {
		if item.ID == target || strings.ToLower(item.Email) == needle || strings.ToLower(item.Alias) == needle {
			return item, nil
		}
	}
	return domain.SessionRecord{}, fmt.Errorf("session %q not found", target)
}

func (s *Store) ListSessions() ([]domain.SessionRecord, string, error) {
	state, err := s.LoadState()
	if err != nil {
		return nil, "", err
	}
	return state.Sessions, state.ActiveSessionID, nil
}
