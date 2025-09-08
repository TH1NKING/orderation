package mysql

import (
    "database/sql"
    "errors"
    "strings"
    "time"

    "orderation/internal/models"
    mem "orderation/internal/store/memory"
)

type UserStore struct { db *sql.DB }

func NewUserStore(db *sql.DB) *UserStore { return &UserStore{db: db} }

func (s *UserStore) Create(u *models.User) error {
    if u.ID == "" { u.ID = mem.NewIDForExternal() }
    if u.CreatedAt.IsZero() { u.CreatedAt = time.Now() }
    _, err := s.db.Exec(`INSERT INTO users (id,name,email,pass_hash,role,created_at) VALUES (?,?,?,?,?,?)`, u.ID, u.Name, strings.ToLower(u.Email), u.PassHash, u.Role, u.CreatedAt)
    if err != nil { return err }
    return nil
}

func (s *UserStore) ByEmail(email string) (*models.User, error) {
    row := s.db.QueryRow(`SELECT id,name,email,pass_hash,role,created_at FROM users WHERE email=?`, strings.ToLower(email))
    var u models.User
    if err := row.Scan(&u.ID,&u.Name,&u.Email,&u.PassHash,&u.Role,&u.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
        return nil, err
    }
    return &u, nil
}

func (s *UserStore) ByID(id string) (*models.User, error) {
    row := s.db.QueryRow(`SELECT id,name,email,pass_hash,role,created_at FROM users WHERE id=?`, id)
    var u models.User
    if err := row.Scan(&u.ID,&u.Name,&u.Email,&u.PassHash,&u.Role,&u.CreatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, errors.New("not found") }
        return nil, err
    }
    return &u, nil
}

