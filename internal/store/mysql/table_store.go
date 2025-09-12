package mysql

import (
    "database/sql"
    "errors"
    "time"

    "orderation/internal/models"
    mem "orderation/internal/store/memory"
)

type TableStore struct { db *sql.DB }

func NewTableStore(db *sql.DB) *TableStore { return &TableStore{db: db} }

func (s *TableStore) Create(t *models.Table) error {
    if t.ID == "" { t.ID = mem.NewIDForExternal() }
    if t.CreatedAt.IsZero() { t.CreatedAt = time.Now() }
    _, err := s.db.Exec(`INSERT INTO tables (id,restaurant_id,name,capacity,created_at) VALUES (?,?,?,?,?)`, t.ID, t.RestaurantID, t.Name, t.Capacity, t.CreatedAt)
    return err
}

func (s *TableStore) ListByRestaurant(restaurantID string) ([]*models.Table, error) {
    rows, err := s.db.Query(`SELECT id,restaurant_id,name,capacity,created_at FROM tables WHERE restaurant_id=? ORDER BY capacity ASC, created_at ASC`, restaurantID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []*models.Table
    for rows.Next() {
        var t models.Table
        if err := rows.Scan(&t.ID,&t.RestaurantID,&t.Name,&t.Capacity,&t.CreatedAt); err != nil { return nil, err }
        out = append(out, &t)
    }
    return out, nil
}

func (s *TableStore) ByID(id string) (*models.Table, error) {
    row := s.db.QueryRow(`SELECT id,restaurant_id,name,capacity,created_at FROM tables WHERE id=?`, id)
    var t models.Table
    if err := row.Scan(&t.ID,&t.RestaurantID,&t.Name,&t.Capacity,&t.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
        return nil, err
    }
    return &t, nil
}

