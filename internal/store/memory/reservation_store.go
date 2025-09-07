package memory

import (
    "errors"
    "sort"
    "sync"
    "time"

    "orderation/internal/models"
    "orderation/internal/store"
)

type ReservationStore struct {
    mu     sync.RWMutex
    byID   map[string]*models.Reservation
    byUser map[string][]string
    byTab  map[string][]string
}

func NewReservationStore() *ReservationStore {
    return &ReservationStore{byID: map[string]*models.Reservation{}, byUser: map[string][]string{}, byTab: map[string][]string{}}
}

func (s *ReservationStore) Create(r *models.Reservation) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    if r.ID == "" {
        r.ID = newID()
    }
    r.CreatedAt = time.Now()
    if r.Status == "" {
        r.Status = "confirmed"
    }
    s.byID[r.ID] = r
    s.byUser[r.UserID] = append(s.byUser[r.UserID], r.ID)
    s.byTab[r.TableID] = append(s.byTab[r.TableID], r.ID)
    return nil
}

func (s *ReservationStore) ByID(id string) (*models.Reservation, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    r := s.byID[id]
    if r == nil {
        return nil, errors.New("not found")
    }
    return r, nil
}

func (s *ReservationStore) Cancel(id string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    r := s.byID[id]
    if r == nil {
        return errors.New("not found")
    }
    r.Status = "cancelled"
    return nil
}

func (s *ReservationStore) ListByUser(userID string) ([]*models.Reservation, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    ids := s.byUser[userID]
    out := make([]*models.Reservation, 0, len(ids))
    for _, id := range ids {
        if r := s.byID[id]; r != nil {
            out = append(out, r)
        }
    }
    sort.Slice(out, func(i, j int) bool { return out[i].StartTime.Before(out[j].StartTime) })
    return out, nil
}

func (s *ReservationStore) ListOverlap(f store.ReservationFilter) ([]*models.Reservation, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    var out []*models.Reservation
    for _, r := range s.byID {
        if f.RestaurantID != "" && r.RestaurantID != f.RestaurantID {
            continue
        }
        if f.TableID != "" && r.TableID != f.TableID {
            continue
        }
        if f.UserID != "" && r.UserID != f.UserID {
            continue
        }
        if r.Status == "cancelled" {
            continue
        }
        // overlap if (r.Start < f.EndAfter) && (r.End > f.StartBefore)
        if r.StartTime.Before(f.StartBefore) && !r.EndTime.After(f.EndAfter) {
            // fully inside window; still overlap
            out = append(out, r)
            continue
        }
        if r.StartTime.Before(f.EndAfter) && r.EndTime.After(f.StartBefore) {
            out = append(out, r)
            continue
        }
    }
    sort.Slice(out, func(i, j int) bool { return out[i].StartTime.Before(out[j].StartTime) })
    return out, nil
}

