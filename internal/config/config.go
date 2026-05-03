// Package config loads and validates application configuration.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config is the top-level application configuration.
type Config struct {
	App       AppConfig       `mapstructure:"app" validate:"required"`
	Server    ServerConfig    `mapstructure:"server" validate:"required"`
	Mongo     MongoConfig     `mapstructure:"mongo" validate:"required"`
	JWT       JWTConfig       `mapstructure:"jwt" validate:"required"`
	Logger    LoggerConfig    `mapstructure:"logger" validate:"required"`
	Notifier  NotifierConfig  `mapstructure:"notifier" validate:"required"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Swagger   SwaggerConfig   `mapstructure:"swagger"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name    string `mapstructure:"name" validate:"required"`
	Env     string `mapstructure:"env" validate:"required,oneof=development staging production"`
	Version string `mapstructure:"version"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port            int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	BodyLimit       string        `mapstructure:"body_limit"`
}

// MongoConfig holds MongoDB connection settings.
type MongoConfig struct {
	URI      string        `mapstructure:"uri" validate:"required"`
	Database string        `mapstructure:"database" validate:"required"`
	MinPool  uint64        `mapstructure:"min_pool"`
	MaxPool  uint64        `mapstructure:"max_pool" validate:"gtefield=MinPool"`
	Timeout  time.Duration `mapstructure:"timeout"`
	AppName  string        `mapstructure:"appname"`
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret string        `mapstructure:"secret" validate:"required,min=32"`
	TTL    time.Duration `mapstructure:"ttl" validate:"required"`
	Issuer string        `mapstructure:"issuer" validate:"required"`
}

// LoggerConfig holds logging settings.
type LoggerConfig struct {
	Level       string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format      string `mapstructure:"format" validate:"required,oneof=json console"`
	Development bool   `mapstructure:"development"`
}

// NotifierConfig holds outbound notification HTTP client settings.
type NotifierConfig struct {
	BaseURL string        `mapstructure:"baseurl" validate:"required,url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Retry   int           `mapstructure:"retry" validate:"min=0,max=10"`
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"alloworigins"`
	AllowCredentials bool          `mapstructure:"allowcredentials"`
	MaxAge           time.Duration `mapstructure:"maxage"`
}

// RateLimitConfig holds rate-limiting settings.
type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	RPS     int  `mapstructure:"rps" validate:"min=0"`
	Burst   int  `mapstructure:"burst" validate:"min=0"`
}

// SwaggerConfig holds Swagger documentation settings.
type SwaggerConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// IsProduction returns true if the application is running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// Load reads configuration from environment variables and validates the result.
func Load() (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config.Load: unmarshal: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config.Load: validation: %w", err)
	}

	return cfg, nil
}

// LoadFromFile reads configuration from a file path, merges with env vars, and validates.
func LoadFromFile(path string) (*Config, error) {
	v := viper.New()

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("config.LoadFromFile: read config: %w", err)
		}
	}

	setDefaults(v)

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config.LoadFromFile: unmarshal: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config.LoadFromFile: validation: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	// App
	v.SetDefault("app.name", "go-starter")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.version", "0.1.0")

	// Server
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)
	v.SetDefault("server.body_limit", "1M")

	// Mongo
	v.SetDefault("mongo.uri", "mongodb://localhost:27017")
	v.SetDefault("mongo.database", "go_starter")
	v.SetDefault("mongo.min_pool", 5)
	v.SetDefault("mongo.max_pool", 100)
	v.SetDefault("mongo.timeout", 10*time.Second)
	v.SetDefault("mongo.appname", "go-starter")

	// JWT
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.ttl", 24*time.Hour)
	v.SetDefault("jwt.issuer", "go-starter")

	// Logger
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("logger.development", false)

	// Notifier
	v.SetDefault("notifier.baseurl", "")
	v.SetDefault("notifier.timeout", 10*time.Second)
	v.SetDefault("notifier.retry", 3)

	// CORS
	v.SetDefault("cors.alloworigins", []string{"http://localhost:3000"})
	v.SetDefault("cors.allowcredentials", false)
	v.SetDefault("cors.maxage", 12*time.Hour)

	// RateLimit
	v.SetDefault("ratelimit.enabled", true)
	v.SetDefault("ratelimit.rps", 20)
	v.SetDefault("ratelimit.burst", 40)

	// Swagger
	v.SetDefault("swagger.enabled", true)
}
