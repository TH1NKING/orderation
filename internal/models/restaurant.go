package models

import "time"

type Restaurant struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Address   string    `json:"address"`
    OpenTime  string    `json:"openTime"`  // e.g., 10:00
    CloseTime string    `json:"closeTime"` // e.g., 22:00
    CreatedAt time.Time `json:"createdAt"`
}

