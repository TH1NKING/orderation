package memory

import (
    "errors"
    "sort"
    "sync"
    "time"

    "orderation/internal/models"
)

type RestaurantStore struct {
    mu   sync.RWMutex
    byID map[string]*models.Restaurant
}

func NewRestaurantStore() *RestaurantStore {
    return &RestaurantStore{byID: map[string]*models.Restaurant{}}
}

func (s *RestaurantStore) Create(r *models.Restaurant) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if r.ID == "" {
        r.ID = newID()
    }
    r.CreatedAt = time.Now()
    s.byID[r.ID] = r
    return nil
}

func (s *RestaurantStore) List() ([]*models.Restaurant, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    out := make([]*models.Restaurant, 0, len(s.byID))
    for _, v := range s.byID {
        out = append(out, v)
    }
    sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
    return out, nil
}

func (s *RestaurantStore) ByID(id string) (*models.Restaurant, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    r := s.byID[id]
    if r == nil {
        return nil, errors.New("not found")
    }
    return r, nil
}

