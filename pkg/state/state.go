package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"trade-desk-newsletter/pkg/models"
)

type Store struct{ path string }

func New(path string) *Store { return &Store{path: path} }

func (s *Store) Load() (*models.State, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// initialize with empty maps
			return &models.State{
				HoldingsByDay: map[string][]string{},
				FirstSeen:     map[string]string{},
				LastSeen:      map[string]string{},
			}, nil
		}
		return nil, err
	}

	var st models.State
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}

	// ensure maps are non-nil
	if st.HoldingsByDay == nil {
		st.HoldingsByDay = map[string][]string{}
	}
	if st.FirstSeen == nil {
		st.FirstSeen = map[string]string{}
	}
	if st.LastSeen == nil {
		st.LastSeen = map[string]string{}
	}

	return &st, nil
}

func (s *Store) Save(st *models.State) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
