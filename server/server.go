package server

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/srechamp/object-store/store"
)

func Run(addr string) error {
    s := store.New()
    h := NewHandler(s)

    mux := http.NewServeMux()
    mux.HandleFunc("PUT /objects/{bucket}/{objectID}", h.Put)
    mux.HandleFunc("GET /objects/{bucket}/{objectID}", h.Get)
    mux.HandleFunc("DELETE /objects/{bucket}/{objectID}", h.Delete)

    srv := &http.Server{
        Addr:         addr,
        Handler:      mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        slog.Info("server starting", "addr", addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "err", err)
            os.Exit(1)
        }
    }()

    <-quit
    slog.Info("shutting down gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return srv.Shutdown(ctx)
}