package server_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
    "time"

    "orderation/internal/server"
)

func TestEndToEndFlow(t *testing.T) {
    os.Setenv("ADMIN_EMAIL", "admin@test.local")
    os.Setenv("ADMIN_PASSWORD", "adminpwd")
    os.Setenv("SECRET", "it-is-a-test-secret")

    srv := server.New()
    ts := httptest.NewServer(srv.Handler())
    defer ts.Close()

    // admin login
    adminTok := login(t, ts.URL, "admin@test.local", "adminpwd")

    // create restaurant
    restBody := map[string]any{"name": "DemoR", "address": "Addr", "openTime": "10:00", "closeTime": "22:00"}
    var restResp map[string]any
    doJSON(t, ts.URL+"/api/v1/restaurants", http.MethodPost, adminTok, restBody, &restResp, 201)
    restID := restResp["id"].(string)

    // create table
    tblBody := map[string]any{"name": "A1", "capacity": 4}
    var tblResp map[string]any
    doJSON(t, ts.URL+"/api/v1/restaurants/"+restID+"/tables", http.MethodPost, adminTok, tblBody, &tblResp, 201)
    tableID := tblResp["id"].(string)

    // register user
    var reg map[string]any
    doJSON(t, ts.URL+"/api/v1/auth/register", http.MethodPost, "", map[string]any{"name": "U", "email": "u@test.local", "password": "p"}, &reg, 201)
    userTok := reg["token"].(string)

    // availability
    start := time.Now().Add(2 * time.Hour).Truncate(time.Minute)
    end := start.Add(2 * time.Hour)
    var avail []map[string]any
    doJSON(t, ts.URL+"/api/v1/restaurants/"+restID+"/availability", http.MethodPost, "", map[string]any{"start": start, "end": end, "guests": 2}, &avail, 200)
    if len(avail) == 0 { t.Fatalf("expected available tables") }

    // create reservation (explicit table)
    var res map[string]any
    doJSON(t, ts.URL+"/api/v1/restaurants/"+restID+"/reservations", http.MethodPost, userTok, map[string]any{"start": start, "end": end, "guests": 2, "tableId": tableID}, &res, 201)
    resID := res["id"].(string)

    // list mine
    var my []map[string]any
    doJSON(t, ts.URL+"/api/v1/me/reservations", http.MethodGet, userTok, nil, &my, 200)
    if len(my) != 1 { t.Fatalf("expected 1 reservation, got %d", len(my)) }

    // cancel
    doJSON(t, ts.URL+"/api/v1/reservations/"+resID, http.MethodDelete, userTok, nil, &res, 200)
    if res["status"].(string) != "cancelled" { t.Fatalf("expected cancelled") }
}

func login(t *testing.T, base, email, password string) string {
    t.Helper()
    var out map[string]any
    doJSON(t, base+"/api/v1/auth/login", http.MethodPost, "", map[string]any{"email": email, "password": password}, &out, 200)
    tok, _ := out["token"].(string)
    if tok == "" { t.Fatalf("missing token in login response") }
    return tok
}

func doJSON(t *testing.T, url, method, token string, body any, out any, want int) {
    t.Helper()
    var buf bytes.Buffer
    if body != nil { if err := json.NewEncoder(&buf).Encode(body); err != nil { t.Fatalf("json encode: %v", err) } }
    req, _ := http.NewRequest(method, url, &buf)
    req.Header.Set("Content-Type", "application/json")
    if token != "" { req.Header.Set("Authorization", "Bearer "+token) }
    resp, err := http.DefaultClient.Do(req)
    if err != nil { t.Fatalf("http: %v", err) }
    defer resp.Body.Close()
    if resp.StatusCode != want { t.Fatalf("%s %s: want %d got %d", method, url, want, resp.StatusCode) }
    if out != nil { _ = json.NewDecoder(resp.Body).Decode(out) }
}

