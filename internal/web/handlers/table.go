package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"

    "orderation/internal/models"
    "orderation/internal/store"
    "orderation/internal/web/router"
)

type TableHandler struct {
    restaurants store.RestaurantStore
    tables      store.TableStore
}

func NewTableHandler(rest store.RestaurantStore, tables store.TableStore) *TableHandler {
    return &TableHandler{restaurants: rest, tables: tables}
}

type createTableReq struct {
    Name     string `json:"name"`
    Capacity int    `json:"capacity"`
}

func (h *TableHandler) Create(w http.ResponseWriter, r *http.Request) {
    rid := router.Param(r, "id")
    if _, err := h.restaurants.ByID(rid); err != nil {
        notFound(w, "restaurant not found")
        return
    }
    var req createTableReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        badRequest(w, "invalid json")
        return
    }
    if req.Capacity <= 0 {
        badRequest(w, "capacity must be > 0")
        return
    }
    t := &models.Table{RestaurantID: rid, Name: req.Name, Capacity: req.Capacity}
    if err := h.tables.Create(t); err != nil {
        badRequest(w, "could not create table")
        return
    }
    writeJSON(w, http.StatusCreated, t)
}

func (h *TableHandler) ListByRestaurant(w http.ResponseWriter, r *http.Request) {
    rid := router.Param(r, "id")
    if _, err := h.restaurants.ByID(rid); err != nil {
        notFound(w, "restaurant not found")
        return
    }
    list, _ := h.tables.ListByRestaurant(rid)
    // optional query filter by min capacity
    if q := r.URL.Query().Get("minCapacity"); q != "" {
        if n, err := strconv.Atoi(q); err == nil {
            out := make([]*models.Table, 0, len(list))
            for _, t := range list {
                if t.Capacity >= n {
                    out = append(out, t)
                }
            }
            list = out
        }
    }
    writeJSON(w, http.StatusOK, list)
}

