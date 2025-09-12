package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/joho/godotenv"
    "orderation/internal/server"
)

func main() {
    // Load .env file if it exists
    if err := godotenv.Load(); err != nil {
        log.Println("[info] no .env file found, using system environment variables")
    } else {
        log.Println("[info] loaded configuration from .env file")
    }

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

