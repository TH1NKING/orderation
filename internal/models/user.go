package models

import "time"

type User struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    PassHash  string    `json:"-"`
    Role      string    `json:"role"` // user | admin
    CreatedAt time.Time `json:"createdAt"`
}

