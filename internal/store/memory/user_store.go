package memory

import (
    "errors"
    "strings"
    "sync"
    "time"

    "orderation/internal/models"
)

type UserStore struct {
    mu    sync.RWMutex
    byID  map[string]*models.User
    byKey map[string]string // lower(email) -> id
}

func NewUserStore() *UserStore {
    return &UserStore{byID: map[string]*models.User{}, byKey: map[string]string{}}
}

func (s *UserStore) Create(u *models.User) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    key := strings.ToLower(u.Email)
    if _, ok := s.byKey[key]; ok {
        return errors.New("email already exists")
    }
    if u.ID == "" {
        u.ID = newID()
    }
    u.CreatedAt = time.Now()
    s.byID[u.ID] = u
    s.byKey[key] = u.ID
    return nil
}

func (s *UserStore) ByEmail(email string) (*models.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    id, ok := s.byKey[strings.ToLower(email)]
    if !ok {
        return nil, errors.New("not found")
    }
    return s.byID[id], nil
}

func (s *UserStore) ByID(id string) (*models.User, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    u := s.byID[id]
    if u == nil {
        return nil, errors.New("not found")
    }
    return u, nil
}

