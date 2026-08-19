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
		Port:          ":8081",
		ReadTimeout:   120 * time.Second, // 2 minutes
		WriteTimeout:  10 * time.Second,
		PingInterval:  45 * time.Second,
		RoundDuration: 90 * time.Second,
	}
}
