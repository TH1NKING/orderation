package handlers

import (
    "encoding/json"
    "net/http"
    "os"
    "strings"
    "time"

    "orderation/internal/auth"
    "orderation/internal/models"
    "orderation/internal/store"
)

type AuthHandler struct {
    users store.UserStore
    pass  *auth.PasswordHasher
    token *auth.TokenManager
}

func NewAuthHandler(users store.UserStore, pass *auth.PasswordHasher, token *auth.TokenManager) *AuthHandler {
    return &AuthHandler{users: users, pass: pass, token: token}
}

type registerReq struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type authResp struct {
    Token string       `json:"token"`
    User  *models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        badRequest(w, "invalid json")
        return
    }
    req.Email = strings.TrimSpace(strings.ToLower(req.Email))
    req.Name = strings.TrimSpace(req.Name)
    if req.Email == "" || req.Password == "" || req.Name == "" {
        badRequest(w, "name, email, password required")
        return
    }
    hash, err := h.pass.Hash(req.Password)
    if err != nil {
        badRequest(w, "unable to hash password")
        return
    }
    u := &models.User{Name: req.Name, Email: req.Email, PassHash: hash, Role: "user"}
    if err := h.users.Create(u); err != nil {
        badRequest(w, "email already exists")
        return
    }
    tok, _ := h.token.Sign(u.ID, u.Role, 24*time.Hour)
    writeJSON(w, http.StatusCreated, authResp{Token: tok, User: &models.User{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, CreatedAt: u.CreatedAt}})
}

type loginReq struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        badRequest(w, "invalid json")
        return
    }
    u, err := h.users.ByEmail(strings.TrimSpace(strings.ToLower(req.Email)))
    if err != nil || !h.pass.Verify(req.Password, u.PassHash) {
        unauthorized(w, "invalid credentials")
        return
    }
    tok, _ := h.token.Sign(u.ID, u.Role, 24*time.Hour)
    writeJSON(w, http.StatusOK, authResp{Token: tok, User: &models.User{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, CreatedAt: u.CreatedAt}})
}

// BootstrapAdmin creates an admin user if ADMIN_EMAIL and ADMIN_PASSWORD are set.
func BootstrapAdmin(users store.UserStore, pass *auth.PasswordHasher) {
    email := strings.TrimSpace(strings.ToLower(os.Getenv("ADMIN_EMAIL")))
    password := os.Getenv("ADMIN_PASSWORD")
    name := os.Getenv("ADMIN_NAME")
    if email == "" || password == "" {
        return
    }
    if name == "" { name = "Admin" }
    if _, err := users.ByEmail(email); err == nil {
        return
    }
    hash, err := pass.Hash(password)
    if err != nil { return }
    _ = users.Create(&models.User{Name: name, Email: email, PassHash: hash, Role: "admin"})
}

