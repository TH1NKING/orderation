package handlers

import (
    "encoding/json"
    "net/http"
    "strings"

    "orderation/internal/models"
    "orderation/internal/store"
    "orderation/internal/web/router"
)

type RestaurantHandler struct {
    stores store.RestaurantStore
}

func NewRestaurantHandler(st store.RestaurantStore) *RestaurantHandler {
    return &RestaurantHandler{stores: st}
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
    if err := h.stores.Create(rest); err != nil {
        badRequest(w, "could not create restaurant")
        return
    }
    writeJSON(w, http.StatusCreated, rest)
}

func (h *RestaurantHandler) List(w http.ResponseWriter, r *http.Request) {
    list, _ := h.stores.List()
    writeJSON(w, http.StatusOK, list)
}

func (h *RestaurantHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    rest, err := h.stores.ByID(id)
    if err != nil {
        notFound(w, "restaurant not found")
        return
    }
    writeJSON(w, http.StatusOK, rest)
}

