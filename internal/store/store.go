package store

import (
    "time"

    "orderation/internal/models"
)

type UserStore interface {
    Create(u *models.User) error
    ByEmail(email string) (*models.User, error)
    ByID(id string) (*models.User, error)
}

type RestaurantStore interface {
    Create(r *models.Restaurant) error
    List() ([]*models.Restaurant, error)
    ByID(id string) (*models.Restaurant, error)
    Delete(id string) error
}

type TableStore interface {
    Create(t *models.Table) error
    ListByRestaurant(restaurantID string) ([]*models.Table, error)
    ByID(id string) (*models.Table, error)
}

type ReservationFilter struct {
    RestaurantID string
    TableID      string
    UserID       string
    StartBefore  time.Time
    EndAfter     time.Time
}

type ReservationStore interface {
    Create(r *models.Reservation) error
    ByID(id string) (*models.Reservation, error)
    Cancel(id string) error
    ListByUser(userID string) ([]*models.Reservation, error)
    ListOverlap(f ReservationFilter) ([]*models.Reservation, error)
}

