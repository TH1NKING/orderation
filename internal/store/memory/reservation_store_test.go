package memory

import (
    "testing"
    "time"

    "orderation/internal/models"
    "orderation/internal/store"
)

func TestListOverlap(t *testing.T) {
    s := NewReservationStore()
    base := time.Now().Truncate(time.Hour)
    r1 := &models.Reservation{RestaurantID: "r1", TableID: "t1", UserID: "u1", StartTime: base.Add(1 * time.Hour), EndTime: base.Add(2 * time.Hour)}
    r2 := &models.Reservation{RestaurantID: "r1", TableID: "t1", UserID: "u2", StartTime: base.Add(3 * time.Hour), EndTime: base.Add(4 * time.Hour)}
    _ = s.Create(r1)
    _ = s.Create(r2)

    // window 1.5h-2.5h overlaps r1 only
    list, _ := s.ListOverlap(store.ReservationFilter{RestaurantID: "r1", TableID: "t1", StartBefore: base.Add(90 * time.Minute), EndAfter: base.Add(150 * time.Minute)})
    if len(list) != 1 || list[0].ID != r1.ID {
        t.Fatalf("expected 1 overlap (r1), got %d", len(list))
    }

    // window 2h-3h no overlap
    list, _ = s.ListOverlap(store.ReservationFilter{RestaurantID: "r1", TableID: "t1", StartBefore: base.Add(2 * time.Hour), EndAfter: base.Add(3 * time.Hour)})
    if len(list) != 0 {
        t.Fatalf("expected 0 overlaps, got %d", len(list))
    }
}

