package models

import "time"

type Reservation struct {
    ID           string    `json:"id"`
    RestaurantID string    `json:"restaurantId"`
    TableID      string    `json:"tableId"`
    UserID       string    `json:"userId"`
    StartTime    time.Time `json:"startTime"`
    EndTime      time.Time `json:"endTime"`
    Guests       int       `json:"guests"`
    Status       string    `json:"status"` // confirmed | cancelled
    CreatedAt    time.Time `json:"createdAt"`
}

