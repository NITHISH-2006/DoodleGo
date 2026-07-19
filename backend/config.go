package main

import "time"

type Config struct {
	Port          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	PingInterval  time.Duration
	RoundDuration time.Duration
}

func DefaultConfig() Config {
	return Config{
		Port:          ":8080",
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  10 * time.Second,
		PingInterval:  30 * time.Second,
		RoundDuration: 90 * time.Second,
	}
}
