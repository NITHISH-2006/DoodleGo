package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	PingInterval   time.Duration
	RoundDuration  time.Duration
	MaxPlayers     int
	MaxMessageSize int64
	AllowedOrigins []string
}

func DefaultConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8081"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	maxPlayers := 20
	if raw := strings.TrimSpace(os.Getenv("MAX_PLAYERS")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 1 && parsed <= 50 {
			maxPlayers = parsed
		}
	}

	origins := []string{"*"}
	if raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")); raw != "" {
		origins = nil
		for _, origin := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	return Config{
		Port:           port,
		ReadTimeout:    120 * time.Second,
		WriteTimeout:   10 * time.Second,
		PingInterval:   45 * time.Second,
		RoundDuration:  90 * time.Second,
		MaxPlayers:     maxPlayers,
		MaxMessageSize: 1 << 20, // 1 MiB
		AllowedOrigins: origins,
	}
}
