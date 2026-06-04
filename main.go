package main

import (
    "log"

    "github.com/srechamp/object-store/config"
    "github.com/srechamp/object-store/server"
)

func main() {
    cfg := config.Load()
    if err := server.Run(cfg.Addr); err != nil {
        log.Fatalf("server exited with error: %v", err)
    }
}