package handlers

import (
    "encoding/json"
    "net/http"
    "strings"
    "time"

    "orderation/internal/models"
    "orderation/internal/store"
    "orderation/internal/web/router"
)

type RestaurantHandler struct {
    restaurants store.RestaurantStore
    tables      store.TableStore
    reservations store.ReservationStore
}

func NewRestaurantHandler(restaurants store.RestaurantStore, tables store.TableStore, reservations store.ReservationStore) *RestaurantHandler {
    return &RestaurantHandler{
        restaurants:  restaurants,
        tables:       tables,
        reservations: reservations,
    }
}

type createRestaurantReq struct {
    Name      string `json:"name"`
    Address   string `json:"address"`
    OpenTime  string `json:"openTime"`
    CloseTime string `json:"closeTime"`
}

func (h *RestaurantHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req createRestaurantReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        badRequest(w, "invalid json")
        return
    }
    req.Name = strings.TrimSpace(req.Name)
    if req.Name == "" {
        badRequest(w, "name required")
        return
    }
    rest := &models.Restaurant{Name: req.Name, Address: strings.TrimSpace(req.Address), OpenTime: strings.TrimSpace(req.OpenTime), CloseTime: strings.TrimSpace(req.CloseTime)}
    if err := h.restaurants.Create(rest); err != nil {
        badRequest(w, "could not create restaurant")
        return
    }
    writeJSON(w, http.StatusCreated, rest)
}

func (h *RestaurantHandler) List(w http.ResponseWriter, r *http.Request) {
    list, _ := h.restaurants.List()
    writeJSON(w, http.StatusOK, list)
}

func (h *RestaurantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    rest, err := h.restaurants.ByID(id)
    if err != nil {
        notFound(w, "restaurant not found")
        return
    }
    writeJSON(w, http.StatusOK, rest)
}

type RestaurantDetails struct {
    *models.Restaurant
    Stats RestaurantStats `json:"stats"`
    Tables []TableInfo     `json:"tables"`
}

type RestaurantStats struct {
    TotalTables       int `json:"totalTables"`
    TotalCapacity     int `json:"totalCapacity"`
    TotalReservations int `json:"totalReservations"`
    ActiveReservations int `json:"activeReservations"`
}

type TableInfo struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Capacity int    `json:"capacity"`
    Status   string `json:"status"` // available, occupied, reserved
}

func (h *RestaurantHandler) GetDetails(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    
    // Get restaurant basic info
    restaurant, err := h.restaurants.ByID(id)
    if err != nil {
        notFound(w, "restaurant not found")
        return
    }
    
    // Get tables info
    tables, _ := h.tables.ListByRestaurant(id)
    var tableInfos []TableInfo
    totalCapacity := 0
    
    for _, table := range tables {
        tableInfo := TableInfo{
            ID:       table.ID,
            Name:     table.Name,
            Capacity: table.Capacity,
            Status:   "available", // Default status
        }
        
        // Check if table has active reservations (simple check for now)
        overlaps, _ := h.reservations.ListOverlap(store.ReservationFilter{
            RestaurantID: id,
            TableID:      table.ID,
            StartBefore:  time.Now(),
            EndAfter:     time.Now(),
        })
        
        if len(overlaps) > 0 {
            tableInfo.Status = "occupied"
        }
        
        tableInfos = append(tableInfos, tableInfo)
        totalCapacity += table.Capacity
    }
    
    // Get reservation statistics (simplified)
    // This could be enhanced with more complex queries
    stats := RestaurantStats{
        TotalTables:       len(tables),
        TotalCapacity:     totalCapacity,
        TotalReservations: 0, // Would need a count query
        ActiveReservations: 0, // Would need a count query for active reservations
    }
    
    details := RestaurantDetails{
        Restaurant: restaurant,
        Stats:      stats,
        Tables:     tableInfos,
    }
    
    writeJSON(w, http.StatusOK, details)
}

func (h *RestaurantHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    
    // Check if restaurant exists
    _, err := h.restaurants.ByID(id)
    if err != nil {
        notFound(w, "restaurant not found")
        return
    }
    
    // Delete the restaurant
    if err := h.restaurants.Delete(id); err != nil {
        badRequest(w, "failed to delete restaurant")
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}

