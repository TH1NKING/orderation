package router

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestRouterParamMatch(t *testing.T) {
    mux := http.NewServeMux()
    r := New(mux)
    r.Handle("GET", "/a/:id/b", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        if Param(req, "id") != "123" {
            t.Fatalf("expected id=123, got %s", Param(req, "id"))
        }
        w.WriteHeader(204)
    }))
    srv := httptest.NewServer(mux)
    defer srv.Close()
    resp, err := http.Get(srv.URL + "/a/123/b")
    if err != nil {
        t.Fatalf("get error: %v", err)
    }
    if resp.StatusCode != 204 {
        t.Fatalf("expected 204, got %d", resp.StatusCode)
    }
}

func TestRouterNotFound(t *testing.T) {
    mux := http.NewServeMux()
    r := New(mux)
    r.Handle("GET", "/ping", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(200) }))
    srv := httptest.NewServer(mux)
    defer srv.Close()
    resp, _ := http.Get(srv.URL + "/unknown")
    if resp.StatusCode != 404 {
        t.Fatalf("expected 404, got %d", resp.StatusCode)
    }
}

