package main

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	PingInterval  time.Duration
	RoundDuration time.Duration
}

func DefaultConfig() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8081"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	return Config{
		Port:          port,
		ReadTimeout:   120 * time.Second, // 2 minutes
		WriteTimeout:  10 * time.Second,
		PingInterval:  45 * time.Second,
		RoundDuration: 90 * time.Second,
	}
}
