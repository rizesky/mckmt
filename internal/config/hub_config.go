package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// HubConfig holds hub-specific configuration
type HubConfig struct {
	Server       ServerConfig       `mapstructure:"server"`
	GRPC         GRPCConfig         `mapstructure:"grpc"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Auth         AuthConfig         `mapstructure:"auth"`
	Orchestrator OrchestratorConfig `mapstructure:"orchestrator"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Metrics      MetricsConfig      `mapstructure:"metrics"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

// LoadHubConfig loads hub configuration from file and environment variables
func LoadHubConfig() (*HubConfig, error) {
	// Reset viper to avoid any previous state
	viper.Reset()

	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/mckmt")

	// Set default values first
	setHubDefaults()

	// Enable reading from environment variables (highest priority)
	viper.SetEnvPrefix("MCKMT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // Enable automatic environment variable binding

	// Check for custom config file from environment variable
	configFile := os.Getenv("MCKMT_CONFIG_FILE")
	if configFile != "" {
		// Use custom config file
		viper.SetConfigFile(configFile)
	} else {
		// Use the default config file name
		viper.SetConfigName("hub_config")
	}

	// Read config file (overrides defaults, but env vars override this)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config HubConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setHubDefaults sets default values for hub configuration
func setHubDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.tls.enabled", false)
	viper.SetDefault("server.tls.cert_file", "")
	viper.SetDefault("server.tls.key_file", "")

	// gRPC defaults
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 8081)
	viper.SetDefault("grpc.read_timeout", "30s")
	viper.SetDefault("grpc.write_timeout", "30s")
	viper.SetDefault("grpc.idle_timeout", "120s")
	viper.SetDefault("grpc.tls.enabled", false)
	viper.SetDefault("grpc.tls.cert_file", "")
	viper.SetDefault("grpc.tls.key_file", "")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "mckmt")
	viper.SetDefault("database.password", "mckmt")
	viper.SetDefault("database.database", "mckmt")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.min_conns", 5)
	viper.SetDefault("database.max_conn_life", "1h")
	viper.SetDefault("database.max_conn_idle", "30m")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Auth defaults
	viper.SetDefault("auth.oidc.enabled", false)
	viper.SetDefault("auth.oidc.issuer", "http://localhost:8080/realms/mckmt")
	viper.SetDefault("auth.oidc.client_id", "mckmt-hub")
	viper.SetDefault("auth.oidc.client_secret", "your-oidc-client-secret")
	viper.SetDefault("auth.oidc.redirect_url", "http://localhost:8080/api/v1/auth/oidc/callback")
	viper.SetDefault("auth.oidc.scopes", []string{"openid", "profile", "email", "groups"})

	viper.SetDefault("auth.jwt.secret", "your-super-secret-jwt-key-change-in-production")
	viper.SetDefault("auth.jwt.expiration", "24h")
	viper.SetDefault("auth.jwt.issuer", "mckmt")
	viper.SetDefault("auth.jwt.audience", "mckmt-users")

	viper.SetDefault("auth.password.enabled", true)
	viper.SetDefault("auth.password.min_length", 8)
	viper.SetDefault("auth.password.require_uppercase", true)
	viper.SetDefault("auth.password.require_lowercase", true)
	viper.SetDefault("auth.password.require_numbers", true)
	viper.SetDefault("auth.password.require_special_chars", true)

	// RBAC defaults
	viper.SetDefault("auth.rbac.strategy", "database-rbac")
	viper.SetDefault("auth.rbac.default_role", "viewer")
	viper.SetDefault("auth.rbac.casbin.enabled", true)
	viper.SetDefault("auth.rbac.casbin.model_file", "configs/casbin-model.conf")
	viper.SetDefault("auth.rbac.casbin.policy_file", "")
	viper.SetDefault("auth.rbac.casbin.auto_reload", false)
	viper.SetDefault("auth.rbac.casbin.reload_interval", 60)

	// Orchestrator defaults
	viper.SetDefault("orchestrator.workers", 5)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.caller", false)
	viper.SetDefault("logging.stacktrace", false)
	viper.SetDefault("logging.output_paths", []string{"stdout"})
	viper.SetDefault("logging.error_output_paths", []string{"stderr"})

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.port", 9091)
}

// Addr returns the server address
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Addr returns the gRPC server address
func (c *GRPCConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
