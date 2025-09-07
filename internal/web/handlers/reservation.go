package handlers

import (
    "encoding/json"
    "net/http"
    "sort"
    "time"

    "orderation/internal/models"
    "orderation/internal/store"
    "orderation/internal/web/middleware"
    "orderation/internal/web/router"
)

type ReservationHandler struct {
    reservations store.ReservationStore
    restaurants  store.RestaurantStore
    tables       store.TableStore
    users        store.UserStore
}

func NewReservationHandler(res store.ReservationStore, rest store.RestaurantStore, tables store.TableStore, users store.UserStore) *ReservationHandler {
    return &ReservationHandler{reservations: res, restaurants: rest, tables: tables, users: users}
}

type availabilityReq struct {
    Start  time.Time `json:"start"`
    End    time.Time `json:"end"`
    Guests int       `json:"guests"`
}

type availabilityResp struct {
    TableID  string `json:"tableId"`
    Capacity int    `json:"capacity"`
}

func (h *ReservationHandler) Availability(w http.ResponseWriter, r *http.Request) {
    rid := router.Param(r, "id")
    if _, err := h.restaurants.ByID(rid); err != nil {
        notFound(w, "restaurant not found")
        return
    }
    var req availabilityReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        badRequest(w, "invalid json")
        return
    }
    if !req.End.After(req.Start) || req.Guests <= 0 {
        badRequest(w, "invalid time range or guests")
        return
    }
    tables, _ := h.tables.ListByRestaurant(rid)
    // sort by capacity asc
    sort.Slice(tables, func(i, j int) bool { return tables[i].Capacity < tables[j].Capacity })
    available := []availabilityResp{}
    for _, t := range tables {
        if t.Capacity < req.Guests {
            continue
        }
        overlaps, _ := h.reservations.ListOverlap(store.ReservationFilter{
            RestaurantID: rid,
            TableID:      t.ID,
            StartBefore:  req.Start,
            EndAfter:     req.End,
        })
        if len(overlaps) == 0 {
            available = append(available, availabilityResp{TableID: t.ID, Capacity: t.Capacity})
        }
    }
    writeJSON(w, http.StatusOK, available)
}

type createReservationReq struct {
    Start  time.Time `json:"start"`
    End    time.Time `json:"end"`
    Guests int       `json:"guests"`
    Table  string    `json:"tableId"`
}

func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
    rid := router.Param(r, "id")
    if _, err := h.restaurants.ByID(rid); err != nil {
        notFound(w, "restaurant not found")
        return
    }
    var req createReservationReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        badRequest(w, "invalid json")
        return
    }
    if !req.End.After(req.Start) || req.Guests <= 0 {
        badRequest(w, "invalid time range or guests")
        return
    }
    // pick table if not provided
    var table *models.Table
    if req.Table != "" {
        t, err := h.tables.ByID(req.Table)
        if err != nil || t.RestaurantID != rid {
            badRequest(w, "invalid tableId")
            return
        }
        table = t
    } else {
        tables, _ := h.tables.ListByRestaurant(rid)
        sort.Slice(tables, func(i, j int) bool { return tables[i].Capacity < tables[j].Capacity })
        for _, t := range tables {
            if t.Capacity < req.Guests {
                continue
            }
            overlaps, _ := h.reservations.ListOverlap(store.ReservationFilter{RestaurantID: rid, TableID: t.ID, StartBefore: req.Start, EndAfter: req.End})
            if len(overlaps) == 0 {
                table = t
                break
            }
        }
        if table == nil {
            badRequest(w, "no available table for the requested time")
            return
        }
    }
    // ensure it's actually available
    overlaps, _ := h.reservations.ListOverlap(store.ReservationFilter{RestaurantID: rid, TableID: table.ID, StartBefore: req.Start, EndAfter: req.End})
    if len(overlaps) > 0 || table.Capacity < req.Guests {
        badRequest(w, "table not available")
        return
    }
    claims := middleware.ClaimsFromContext(r)
    if claims == nil {
        unauthorized(w, "no auth")
        return
    }
    res := &models.Reservation{RestaurantID: rid, TableID: table.ID, UserID: claims.Sub, StartTime: req.Start, EndTime: req.End, Guests: req.Guests, Status: "confirmed"}
    if err := h.reservations.Create(res); err != nil {
        badRequest(w, "could not create reservation")
        return
    }
    writeJSON(w, http.StatusCreated, res)
}

func (h *ReservationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    claims := middleware.ClaimsFromContext(r)
    if claims == nil {
        unauthorized(w, "no auth")
        return
    }
    res, err := h.reservations.ByID(id)
    if err != nil {
        notFound(w, "reservation not found")
        return
    }
    if claims.Role != "admin" && res.UserID != claims.Sub {
        forbidden(w, "not allowed")
        return
    }
    if err := h.reservations.Cancel(id); err != nil {
        badRequest(w, "unable to cancel")
        return
    }
    writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *ReservationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
    claims := middleware.ClaimsFromContext(r)
    if claims == nil {
        unauthorized(w, "no auth")
        return
    }
    list, _ := h.reservations.ListByUser(claims.Sub)
    writeJSON(w, http.StatusOK, list)
}

