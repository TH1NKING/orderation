package models

import "time"

type Table struct {
    ID           string    `json:"id"`
    RestaurantID string    `json:"restaurantId"`
    Name         string    `json:"name"`
    Capacity     int       `json:"capacity"`
    CreatedAt    time.Time `json:"createdAt"`
}

