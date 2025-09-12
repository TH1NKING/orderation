package router

import (
    "context"
    "net/http"
    "strings"
)

type key int

const paramsKey key = 1

type route struct {
    method   string
    pattern  string
    segments []segment
    handler  http.Handler
}

type segment struct {
    name     string // if empty and wildcard == false, it's a literal value in name
    wildcard bool
}

type Router struct {
    routes []route
}

func New(mux *http.ServeMux) *Router {
    r := &Router{}
    mux.Handle("/", r)
    return r
}

func (r *Router) Handle(method, pattern string, handler http.Handler) {
    segs := parsePattern(pattern)
    r.routes = append(r.routes, route{method: method, pattern: pattern, segments: segs, handler: handler})
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Set CORS headers for all requests
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    
    // Handle preflight requests
    if req.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    
    path := req.URL.Path
    for _, rt := range r.routes {
        if req.Method != rt.method {
            continue
        }
        params, ok := match(rt.segments, path)
        if !ok {
            continue
        }
        // attach params to context
        ctx := context.WithValue(req.Context(), paramsKey, params)
        rt.handler.ServeHTTP(w, req.WithContext(ctx))
        return
    }
    http.NotFound(w, req)
}

func parsePattern(pattern string) []segment {
    parts := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
    segs := make([]segment, 0, len(parts))
    for _, p := range parts {
        if p == "" {
            continue
        }
        if strings.HasPrefix(p, ":") {
            name := strings.TrimPrefix(p, ":")
            if name == "" {
                name = "param"
            }
            segs = append(segs, segment{name: name, wildcard: true})
        } else {
            segs = append(segs, segment{name: p, wildcard: false})
        }
    }
    return segs
}

func match(segs []segment, path string) (map[string]string, bool) {
    parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
    // Remove empty from trailing slashes
    compact := make([]string, 0, len(parts))
    for _, p := range parts {
        if p != "" {
            compact = append(compact, p)
        }
    }
    parts = compact
    if len(parts) != len(segs) {
        return nil, false
    }
    params := map[string]string{}
    for i, s := range segs {
        if s.wildcard {
            params[s.name] = parts[i]
        } else if s.name != parts[i] {
            return nil, false
        }
    }
    return params, true
}

// Param returns named path param from request context.
func Param(r *http.Request, name string) string {
    v := r.Context().Value(paramsKey)
    if v == nil {
        return ""
    }
    m, _ := v.(map[string]string)
    return m[name]
}

