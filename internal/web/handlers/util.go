package handlers

import (
    "encoding/json"
    "net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if v != nil {
        _ = json.NewEncoder(w).Encode(v)
    }
}

func badRequest(w http.ResponseWriter, msg string) {
    writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func unauthorized(w http.ResponseWriter, msg string) {
    writeJSON(w, http.StatusUnauthorized, map[string]string{"error": msg})
}

func forbidden(w http.ResponseWriter, msg string) {
    writeJSON(w, http.StatusForbidden, map[string]string{"error": msg})
}

func notFound(w http.ResponseWriter, msg string) {
    writeJSON(w, http.StatusNotFound, map[string]string{"error": msg})
}

