package handlers

import (
    "encoding/json"
    "net/http"
    "sort"
    "strconv"
    "strings"
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

// isWithinOperatingHours checks if the reservation time is within restaurant operating hours
func (h *ReservationHandler) isWithinOperatingHours(restaurant *models.Restaurant, start, end time.Time) bool {
    // Parse operating hours (format: "09:00")
    openHour, openMin, err := parseTime(restaurant.OpenTime)
    if err != nil {
        return false
    }
    closeHour, closeMin, err := parseTime(restaurant.CloseTime)
    if err != nil {
        return false
    }
    
    // Convert UTC times to local time (assuming restaurant operates in Asia/Shanghai timezone)
    loc, err := time.LoadLocation("Asia/Shanghai")
    if err != nil {
        // Fallback to UTC+8 if timezone loading fails
        loc = time.FixedZone("CST", 8*3600)
    }
    
    localStart := start.In(loc)
    localEnd := end.In(loc)
    
    // Get the date and time components in local time
    startDate := localStart.Truncate(24 * time.Hour)
    endDate := localEnd.Truncate(24 * time.Hour)
    
    // Check each day of the reservation in local time
    for date := startDate; !date.After(endDate); date = date.Add(24 * time.Hour) {
        // Create operating hours for this specific date in the local timezone
        openTime := time.Date(date.Year(), date.Month(), date.Day(), openHour, openMin, 0, 0, loc)
        closeTime := time.Date(date.Year(), date.Month(), date.Day(), closeHour, closeMin, 0, 0, loc)
        
        // Handle overnight hours (e.g., 22:00 - 02:00)
        if closeTime.Before(openTime) {
            closeTime = closeTime.Add(24 * time.Hour)
        }
        
        // Check if reservation overlaps with this day's operating hours
        dayStart := localStart
        if localStart.Before(date) {
            dayStart = date
        }
        dayEnd := localEnd
        if localEnd.After(date.Add(24*time.Hour)) {
            dayEnd = date.Add(24 * time.Hour)
        }
        
        // If there's any part of the reservation on this day
        if dayStart.Before(dayEnd) {
            // Check if this part is within operating hours
            if dayStart.Before(openTime) || dayEnd.After(closeTime) {
                return false // Any part outside operating hours means rejection
            }
        }
    }
    
    return true
}

// findBestAvailableTable finds the most suitable available table using smart allocation
func (h *ReservationHandler) findBestAvailableTable(restaurantID string, start, end time.Time, guests int) *models.Table {
    tables, err := h.tables.ListByRestaurant(restaurantID)
    if err != nil {
        return nil
    }
    
    var availableTables []*models.Table
    
    // First, find all available tables that can accommodate the guests
    for _, t := range tables {
        if t.Capacity < guests {
            continue
        }
        
        overlaps, _ := h.reservations.ListOverlap(store.ReservationFilter{
            RestaurantID: restaurantID,
            TableID:      t.ID,
            StartBefore:  start,
            EndAfter:     end,
        })
        
        if len(overlaps) == 0 {
            availableTables = append(availableTables, t)
        }
    }
    
    if len(availableTables) == 0 {
        return nil
    }
    
    // Smart allocation strategy:
    // 1. Prefer tables with capacity closest to guest count (minimize waste)
    // 2. If multiple tables have same capacity, choose randomly
    sort.Slice(availableTables, func(i, j int) bool {
        // Sort by capacity difference from guest count (ascending)
        diffI := availableTables[i].Capacity - guests
        diffJ := availableTables[j].Capacity - guests
        if diffI != diffJ {
            return diffI < diffJ
        }
        // If same difference, sort by table ID for deterministic behavior
        return availableTables[i].ID < availableTables[j].ID
    })
    
    return availableTables[0]
}

// parseTime parses time string in "HH:MM" format
func parseTime(timeStr string) (hour, minute int, err error) {
    parts := strings.Split(timeStr, ":")
    if len(parts) != 2 {
        return 0, 0, err
    }
    
    hour, err = strconv.Atoi(parts[0])
    if err != nil {
        return 0, 0, err
    }
    
    minute, err = strconv.Atoi(parts[1])
    if err != nil {
        return 0, 0, err
    }
    
    return hour, minute, nil
}

type createReservationReq struct {
    Start  time.Time `json:"start"`
    End    time.Time `json:"end"`
    Guests int       `json:"guests"`
    Table  string    `json:"tableId"`
}

func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
    rid := router.Param(r, "id")
    restaurant, err := h.restaurants.ByID(rid)
    if err != nil {
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
    
    // Check if reservation time is within restaurant operating hours
    if !h.isWithinOperatingHours(restaurant, req.Start, req.End) {
        badRequest(w, "reservation time is outside restaurant operating hours")
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
        // Smart table allocation: find the best available table
        table = h.findBestAvailableTable(rid, req.Start, req.End, req.Guests)
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

