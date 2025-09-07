package middleware

import (
    "context"
    "net/http"
    "strings"

    "orderation/internal/auth"
)

type ctxKey int

const claimsKey ctxKey = 1

func RequireAuth(tm *auth.TokenManager, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := parseAuthHeader(r.Header.Get("Authorization"))
        if token == "" {
            http.Error(w, "missing bearer token", http.StatusUnauthorized)
            return
        }
        c, err := tm.Verify(token)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), claimsKey, c)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func RequireRole(tm *auth.TokenManager, role string, next http.Handler) http.Handler {
    return RequireAuth(tm, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        c := ClaimsFromContext(r)
        if c == nil || c.Role != role {
            http.Error(w, "forbidden", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    }))
}

func ClaimsFromContext(r *http.Request) *auth.Claims {
    v := r.Context().Value(claimsKey)
    if v == nil {
        return nil
    }
    c, _ := v.(*auth.Claims)
    return c
}

func parseAuthHeader(h string) string {
    if h == "" {
        return ""
    }
    parts := strings.SplitN(h, " ", 2)
    if len(parts) != 2 {
        return ""
    }
    if !strings.EqualFold(parts[0], "Bearer") {
        return ""
    }
    return strings.TrimSpace(parts[1])
}

