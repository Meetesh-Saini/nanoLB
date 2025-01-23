package config

import (
	"sync"
	"time"
)

type Config struct {
	SessionTimeout            time.Duration
	MaxAttempts               int
	MaxRetries                int
	Algorithm                 ALGORITHM_TYPE
	HealthCheckTimeout        time.Duration
	MaxConcurrentHealthChecks int
	HealthCheckInterval       time.Duration
}

var (
	configInstance *Config
	once           sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		configInstance = &Config{
			SessionTimeout:            10 * time.Minute,
			MaxAttempts:               3,
			MaxRetries:                3,
			HealthCheckTimeout:        2 * time.Second,
			MaxConcurrentHealthChecks: 256,
			HealthCheckInterval:       10 * time.Second,
		}
	})
	return configInstance
}
