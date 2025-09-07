package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "orderation/internal/server"
)

func main() {
    addr := getEnv("ADDR", ":8080")

    srv := server.New()

    httpServer := &http.Server{
        Addr:              addr,
        Handler:           srv.Handler(),
        ReadTimeout:       15 * time.Second,
        WriteTimeout:      15 * time.Second,
        ReadHeaderTimeout: 5 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    go func() {
        log.Printf("server listening on %s", addr)
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit
    log.Println("shutting down server...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := httpServer.Shutdown(ctx); err != nil {
        log.Fatalf("server forced to shutdown: %v", err)
    }
    log.Println("server exited cleanly")
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}

