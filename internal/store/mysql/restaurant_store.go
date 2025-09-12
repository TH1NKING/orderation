package mysql

import (
    "database/sql"
    "errors"
    "time"

    "orderation/internal/models"
    mem "orderation/internal/store/memory"
)

type RestaurantStore struct { db *sql.DB }

func NewRestaurantStore(db *sql.DB) *RestaurantStore { return &RestaurantStore{db: db} }

func (s *RestaurantStore) Create(r *models.Restaurant) error {
    if r.ID == "" { r.ID = mem.NewIDForExternal() }
    if r.CreatedAt.IsZero() { r.CreatedAt = time.Now() }
    _, err := s.db.Exec(`INSERT INTO restaurants (id,name,address,open_time,close_time,created_at) VALUES (?,?,?,?,?,?)`, r.ID, r.Name, r.Address, r.OpenTime, r.CloseTime, r.CreatedAt)
    return err
}

func (s *RestaurantStore) List() ([]*models.Restaurant, error) {
    rows, err := s.db.Query(`SELECT id,name,address,open_time,close_time,created_at FROM restaurants ORDER BY created_at ASC, id ASC`)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []*models.Restaurant
    for rows.Next() {
        var r models.Restaurant
        if err := rows.Scan(&r.ID,&r.Name,&r.Address,&r.OpenTime,&r.CloseTime,&r.CreatedAt); err != nil { return nil, err }
        out = append(out, &r)
    }
    return out, nil
}

func (s *RestaurantStore) ByID(id string) (*models.Restaurant, error) {
    row := s.db.QueryRow(`SELECT id,name,address,open_time,close_time,created_at FROM restaurants WHERE id=?`, id)
    var r models.Restaurant
    if err := row.Scan(&r.ID,&r.Name,&r.Address,&r.OpenTime,&r.CloseTime,&r.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
        return nil, err
    }
    return &r, nil
}

func (s *RestaurantStore) Delete(id string) error {
    result, err := s.db.Exec(`DELETE FROM restaurants WHERE id=?`, id)
    if err != nil {
        return err
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return errors.New("restaurant not found")
    }
    return nil
}

