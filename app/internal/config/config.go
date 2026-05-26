package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var validate = validator.New()

type AppConfig struct {
	ID            string `mapstructure:"id"            validate:"required"`
	Name          string `mapstructure:"name"          validate:"required"`
	Version       string `mapstructure:"version"`
	IsDevelopment bool   `mapstructure:"is_dev"`
	LogLevel      string `mapstructure:"log_level"     validate:"omitempty,oneof=debug info warn error"`
	IsLogJSON     bool   `mapstructure:"is_log_json"`
	Domain        string `mapstructure:"domain"`
}

type GRPCConfig struct {
	Host string `mapstructure:"host" validate:"required"`
	Port int    `mapstructure:"port" validate:"required,gt=0,lt=65536"`
}

type HTTPConfig struct {
	Host              string        `mapstructure:"host"                validate:"required"`
	Port              int           `mapstructure:"port"                validate:"required,gt=0,lt=65536"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
}

type PostgresConfig struct {
	Host       string        `mapstructure:"host"        validate:"required"`
	User       string        `mapstructure:"user"        validate:"required"`
	Password   string        `mapstructure:"password"    validate:"required"`
	Port       int           `mapstructure:"port"        validate:"required,gt=0,lt=65536"`
	Database   string        `mapstructure:"database"    validate:"required"`
	MaxAttempt int           `mapstructure:"max_attempt" validate:"omitempty,gte=1"`
	MaxDelay   time.Duration `mapstructure:"max_delay"`
	RequireSSL bool          `mapstructure:"require_ssl"`
}

type TracingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port" validate:"omitempty,gt=0"`
}

type MetricsConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port" validate:"omitempty,gt=0,lt=65536"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	Enabled           bool          `mapstructure:"enabled"`
}

type ProfilerConfig struct {
	IsEnabled         bool          `mapstructure:"enabled"`
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port" validate:"omitempty,gt=0"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
}

type WebSocketConfig struct {
	ReconnectTimes uint8         `mapstructure:"reconnect_times"`
	ReconnectDelay time.Duration `mapstructure:"reconnect_delay"`
}

type Config struct {
	App       AppConfig       `mapstructure:"app"       validate:"required"`
	GRPC      GRPCConfig      `mapstructure:"grpc"      validate:"required"` // ← добавь
	HTTP      HTTPConfig      `mapstructure:"http"      validate:"required"`
	Postgres  PostgresConfig  `mapstructure:"postgres"  validate:"required"`
	Tracing   TracingConfig   `mapstructure:"tracing"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Profiler  ProfilerConfig  `mapstructure:"profiler"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
}

var globalConfig atomic.Pointer[Config]

func Get() *Config {
	return globalConfig.Load()
}

func MustLoadConfig(configPath string) *Config {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	return cfg
}

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		if envPath, ok := os.LookupEnv("CONFIG_PATH"); ok {
			configPath = envPath
		} else {
			return nil, fmt.Errorf("config path is not provided")
		}
	}

	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(absPath)
	v.SetConfigType("yaml")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig.Store(&cfg)
	slog.Info("config loaded successfully", "path", absPath)

	// Настройка горячего обновления
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		slog.Info("config file changed, reloading...", "file", e.Name)

		var newCfg Config
		if err := v.Unmarshal(&newCfg); err != nil {
			slog.Error("failed to unmarshal config on reload", "error", err)
			return
		}

		if err := validateConfig(&newCfg); err != nil {
			slog.Error("new config is invalid, keeping old config", "error", err)
			return
		}

		globalConfig.Store(&newCfg)
		slog.Info("config reloaded successfully")
	})

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}
