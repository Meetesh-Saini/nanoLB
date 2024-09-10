package config

import (
	"sync"
	"time"
)

type Config struct {
	SessionTimeout time.Duration
	MaxAttempts    int
	MaxRetries     int
	Algorithm      ALGORITHM_TYPE
}

var (
	configInstance *Config
	once           sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		configInstance = &Config{
			SessionTimeout: 10 * time.Minute,
			MaxAttempts:    1,
			MaxRetries:     1,
		}
	})
	return configInstance
}
