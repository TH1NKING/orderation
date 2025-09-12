package memory

import (
    "errors"
    "sort"
    "sync"
    "time"

    "orderation/internal/models"
)

type TableStore struct {
    mu            sync.RWMutex
    byID          map[string]*models.Table
    byRestaurant  map[string][]string
}

func NewTableStore() *TableStore {
    return &TableStore{byID: map[string]*models.Table{}, byRestaurant: map[string][]string{}}
}

func (s *TableStore) Create(t *models.Table) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if t.ID == "" {
        t.ID = newID()
    }
    t.CreatedAt = time.Now()
    s.byID[t.ID] = t
    s.byRestaurant[t.RestaurantID] = append(s.byRestaurant[t.RestaurantID], t.ID)
    return nil
}

func (s *TableStore) ListByRestaurant(restaurantID string) ([]*models.Table, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    ids := s.byRestaurant[restaurantID]
    out := make([]*models.Table, 0, len(ids))
    for _, id := range ids {
        if t := s.byID[id]; t != nil {
            out = append(out, t)
        }
    }
    sort.Slice(out, func(i, j int) bool { return out[i].Capacity < out[j].Capacity })
    return out, nil
}

func (s *TableStore) ByID(id string) (*models.Table, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    t := s.byID[id]
    if t == nil {
        return nil, errors.New("not found")
    }
    return t, nil
}

