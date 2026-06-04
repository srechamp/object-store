package config

import (
    "flag"
    "fmt"
)

type Config struct {
    Addr string
}

func Load() Config {
    port := flag.Int("port", 8080, "HTTP server port")
    flag.Parse()
    return Config{Addr: fmt.Sprintf(":%d", *port)}
}