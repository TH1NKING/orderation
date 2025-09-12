package mysql

import (
    "database/sql"
    "errors"
    "time"

    "orderation/internal/models"
    "orderation/internal/store"
    mem "orderation/internal/store/memory"
)

type ReservationStore struct { db *sql.DB }

func NewReservationStore(db *sql.DB) *ReservationStore { return &ReservationStore{db: db} }

func (s *ReservationStore) Create(r *models.Reservation) error {
    if r.ID == "" { r.ID = mem.NewIDForExternal() }
    if r.CreatedAt.IsZero() { r.CreatedAt = time.Now() }
    if r.Status == "" { r.Status = "confirmed" }
    _, err := s.db.Exec(`INSERT INTO reservations (id,restaurant_id,table_id,user_id,start_time,end_time,guests,status,created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
        r.ID, r.RestaurantID, r.TableID, r.UserID, r.StartTime, r.EndTime, r.Guests, r.Status, r.CreatedAt)
    return err
}

func (s *ReservationStore) ByID(id string) (*models.Reservation, error) {
    row := s.db.QueryRow(`SELECT id,restaurant_id,table_id,user_id,start_time,end_time,guests,status,created_at FROM reservations WHERE id=?`, id)
    var r models.Reservation
    if err := row.Scan(&r.ID,&r.RestaurantID,&r.TableID,&r.UserID,&r.StartTime,&r.EndTime,&r.Guests,&r.Status,&r.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
        return nil, err
    }
    return &r, nil
}

func (s *ReservationStore) Cancel(id string) error {
    _, err := s.db.Exec(`UPDATE reservations SET status='cancelled' WHERE id=?`, id)
    return err
}

func (s *ReservationStore) ListByUser(userID string) ([]*models.Reservation, error) {
    rows, err := s.db.Query(`SELECT id,restaurant_id,table_id,user_id,start_time,end_time,guests,status,created_at FROM reservations WHERE user_id=? ORDER BY start_time ASC, id ASC`, userID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []*models.Reservation
    for rows.Next() {
        var r models.Reservation
        if err := rows.Scan(&r.ID,&r.RestaurantID,&r.TableID,&r.UserID,&r.StartTime,&r.EndTime,&r.Guests,&r.Status,&r.CreatedAt); err != nil { return nil, err }
        out = append(out, &r)
    }
    // Ensure we always return a slice, not nil
    if out == nil {
        out = []*models.Reservation{}
    }
    return out, nil
}

func (s *ReservationStore) ListOverlap(f store.ReservationFilter) ([]*models.Reservation, error) {
    // Build query with optional filters
    q := `SELECT id,restaurant_id,table_id,user_id,start_time,end_time,guests,status,created_at
          FROM reservations
          WHERE status <> 'cancelled' AND start_time < ? AND end_time > ?`
    args := []any{f.EndAfter, f.StartBefore}
    if f.RestaurantID != "" { q += " AND restaurant_id = ?"; args = append(args, f.RestaurantID) }
    if f.TableID != "" { q += " AND table_id = ?"; args = append(args, f.TableID) }
    if f.UserID != "" { q += " AND user_id = ?"; args = append(args, f.UserID) }
    q += " ORDER BY start_time ASC, id ASC"
    rows, err := s.db.Query(q, args...)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []*models.Reservation
    for rows.Next() {
        var r models.Reservation
        if err := rows.Scan(&r.ID,&r.RestaurantID,&r.TableID,&r.UserID,&r.StartTime,&r.EndTime,&r.Guests,&r.Status,&r.CreatedAt); err != nil { return nil, err }
        out = append(out, &r)
    }
    return out, nil
}

