package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	GRPC         GRPCConfig         `mapstructure:"grpc"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Auth         AuthConfig         `mapstructure:"auth"`
	Agent        AgentConfig        `mapstructure:"agent"`
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

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
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

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host        string        `mapstructure:"host"`
	Port        int           `mapstructure:"port"`
	User        string        `mapstructure:"user"`
	Password    string        `mapstructure:"password"`
	Database    string        `mapstructure:"database"`
	SSLMode     string        `mapstructure:"ssl_mode"`
	MaxConns    int           `mapstructure:"max_conns"`
	MinConns    int           `mapstructure:"min_conns"`
	MaxConnLife time.Duration `mapstructure:"max_conn_life"`
	MaxConnIdle time.Duration `mapstructure:"max_conn_idle"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	OIDC OIDCConfig `mapstructure:"oidc"`
	JWT  JWTConfig  `mapstructure:"jwt"`
	RBAC RBACConfig `mapstructure:"rbac"`
}

// OIDCConfig holds OIDC configuration
type OIDCConfig struct {
	Issuer       string `mapstructure:"issuer"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	ModelFile  string `mapstructure:"model_file"`
	PolicyFile string `mapstructure:"policy_file"`
}

// AgentConfig holds agent configuration
type AgentConfig struct {
	HubURL        string        `mapstructure:"hub_url"`
	Token         string        `mapstructure:"token"`
	Heartbeat     time.Duration `mapstructure:"heartbeat"`
	ReconnectWait time.Duration `mapstructure:"reconnect_wait"`
}

// OrchestratorConfig holds orchestrator configuration
type OrchestratorConfig struct {
	Workers int `mapstructure:"workers"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	Caller     bool   `mapstructure:"caller"`
	Stacktrace bool   `mapstructure:"stacktrace"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/mckmt")

	// Set default values
	setDefaults()

	// Enable reading from environment variables
	viper.SetEnvPrefix("MCKMT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults sets minimal default configuration values
// Most defaults are now in configs/config.yaml
func setDefaults() {
	// Only set critical defaults that must be available before config file is read
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 8081)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "mckmt")
	viper.SetDefault("database.password", "mckmt")
	viper.SetDefault("database.database", "mckmt")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

// Addr returns the server address
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Addr returns the gRPC server address
func (c *GRPCConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// RedisAddr returns the Redis address
func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
