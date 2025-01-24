package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

var (
	configInstance *Config
	once           sync.Once
	configFilePath string
)

type Server struct {
	URL    string `toml:"url"`
	Weight int64  `toml:"weight"`
}

// Config struct for global configuration
type Config struct {
	SessionTimeout            time.Duration  `toml:"sessionTimeout"`
	MaxAttempts               int            `toml:"maxAttempts"`
	MaxRetries                int            `toml:"maxRetries"`
	Algorithm                 ALGORITHM_TYPE `toml:"algorithm"`
	HealthCheckTimeout        time.Duration  `toml:"healthCheckTimeout"`
	MaxConcurrentHealthChecks int            `toml:"maxConcurrentHealthChecks"`
	HealthCheckInterval       time.Duration  `toml:"healthCheckInterval"`
	LogFile                   string         `toml:"logFile"`
	LogLevel                  string         `toml:"logLevel"`
	LogFormat                 string         `toml:"logFormat"`
	LogOutput                 string         `toml:"logOutput"`
	Servers                   []Server       `toml:"server"`
}

// Custom unmarshaler for Algorithm type
func (a *ALGORITHM_TYPE) UnmarshalTOML(value interface{}) error {
	switch v := value.(type) {
	case string:
		switch strings.ToLower(v) {
		case "round-robin":
			*a = RoundRobin
		case "weighted-round-robin":
			*a = WeightedRoundRobin
		default:
			return fmt.Errorf("invalid algorithm: %s", v)
		}
		return nil
	default:
		return fmt.Errorf("invalid type for ALGORITHM_TYPE: %T", v)
	}
}

func (s *Server) ValidateURL() error {
	parsedURL, err := url.Parse(s.URL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s, must be http or https", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a valid host")
	}

	// Rebuild the URL with the fixed scheme if necessary
	s.URL = parsedURL.String()

	return nil
}

// ParseConfig loads the configuration from a TOML file
func ParseConfig(filePath string) (*Config, error) {
	var DefaultConfig = Config{
		SessionTimeout:            10 * time.Minute,
		MaxAttempts:               3,
		MaxRetries:                3,
		Algorithm:                 RoundRobin,
		HealthCheckTimeout:        5 * time.Second,
		MaxConcurrentHealthChecks: 16,
		HealthCheckInterval:       10 * time.Minute,
		LogFile:                   "nanolb.log",
		LogLevel:                  "info",
		LogFormat:                 "text",
		LogOutput:                 "both",
		Servers:                   []Server{},
	}
	_, err := toml.DecodeFile(filePath, &DefaultConfig)
	if err != nil {
		return nil, err
	}

	for i := range DefaultConfig.Servers {
		s := &DefaultConfig.Servers[i]
		if err := s.ValidateURL(); err != nil {
			log.Fatal(err)
		}
		if s.Weight == 0 {
			s.Weight = 1
		}
	}

	return &DefaultConfig, nil
}

// SetConfigFilePath allows to set the config file path once
func SetConfigFilePath(filePath string) {
	if _, err := os.Stat(filePath); err != nil {
		log.Fatal("No config found at ", filePath)
	}

	if configFilePath == "" {
		configFilePath = filePath
	}
}

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	once.Do(func() {
		if configFilePath == "" {
			log.Fatal("No config path found")
		}

		// Load config from the provided path
		config, err := ParseConfig(configFilePath)
		if err != nil {
			log.Fatal("Error loading config: ", err)
		}

		// Set the loaded config instance
		configInstance = config
	})
	return configInstance
}
