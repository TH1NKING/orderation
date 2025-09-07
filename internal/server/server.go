package server

import (
    "log"
    "net/http"
    "os"

    "orderation/internal/auth"
    h "orderation/internal/web/handlers"
    "orderation/internal/web/middleware"
    "orderation/internal/web/router"
    "orderation/internal/store/memory"
)

type Server struct {
    mux *http.ServeMux
}

func New() *Server {
    mux := http.NewServeMux()

    // Stores (in-memory implementation)
    userStore := memory.NewUserStore()
    restaurantStore := memory.NewRestaurantStore()
    tableStore := memory.NewTableStore()
    reservationStore := memory.NewReservationStore()

    // Auth setup
    secret := os.Getenv("SECRET")
    if secret == "" {
        // Not ideal for production, but fine for demo
        secret = auth.GenerateRandomSecret()
        log.Println("[warn] SECRET not set; generated ephemeral secret. Tokens reset on restart.")
    }
    token := auth.NewTokenManager(secret)
    pass := auth.NewPasswordHasher(200_000) // iterative salted hash (placeholder)

    // Bootstrap admin if env provided
    h.BootstrapAdmin(userStore, pass)

    // Handlers
    ah := h.NewAuthHandler(userStore, pass, token)
    rh := h.NewRestaurantHandler(restaurantStore)
    th := h.NewTableHandler(restaurantStore, tableStore)
    resvh := h.NewReservationHandler(reservationStore, restaurantStore, tableStore, userStore)

    r := router.New(mux)

    // Health
    r.Handle("GET", "/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"ok":true}`))
    }))

    // Auth
    r.Handle("POST", "/api/v1/auth/register", http.HandlerFunc(ah.Register))
    r.Handle("POST", "/api/v1/auth/login", http.HandlerFunc(ah.Login))

    // Restaurants
    r.Handle("GET", "/api/v1/restaurants", http.HandlerFunc(rh.List))
    r.Handle("GET", "/api/v1/restaurants/:id", http.HandlerFunc(rh.GetByID))
    r.Handle("POST", "/api/v1/restaurants", middleware.RequireRole(token, "admin", http.HandlerFunc(rh.Create)))

    // Tables
    r.Handle("GET", "/api/v1/restaurants/:id/tables", http.HandlerFunc(th.ListByRestaurant))
    r.Handle("POST", "/api/v1/restaurants/:id/tables", middleware.RequireRole(token, "admin", http.HandlerFunc(th.Create)))

    // Availability and reservations
    r.Handle("POST", "/api/v1/restaurants/:id/availability", http.HandlerFunc(resvh.Availability))
    r.Handle("POST", "/api/v1/restaurants/:id/reservations", middleware.RequireAuth(token, http.HandlerFunc(resvh.Create)))
    r.Handle("DELETE", "/api/v1/reservations/:id", middleware.RequireAuth(token, http.HandlerFunc(resvh.Cancel)))
    r.Handle("GET", "/api/v1/me/reservations", middleware.RequireAuth(token, http.HandlerFunc(resvh.ListMine)))

    return &Server{mux: mux}
}

func (s *Server) Handler() http.Handler { return s.mux }
