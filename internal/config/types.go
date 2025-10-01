package config

import (
	"fmt"
	"time"
)

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level            string   `mapstructure:"level"`
	Format           string   `mapstructure:"format"`
	Caller           bool     `mapstructure:"caller"`
	Stacktrace       bool     `mapstructure:"stacktrace"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
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
	Enabled      bool     `mapstructure:"enabled"`
	Issuer       string   `mapstructure:"issuer"`
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
	Issuer     string        `mapstructure:"issuer"`
	Audience   string        `mapstructure:"audience"`
}

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	Strategy    string       `mapstructure:"strategy"` // "no-auth", "database-rbac", "casbin"
	DefaultRole string       `mapstructure:"default_role"`
	Casbin      CasbinConfig `mapstructure:"casbin"`
}

// CasbinConfig holds Casbin-specific configuration
type CasbinConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	ModelFile      string `mapstructure:"model_file"`
	PolicyFile     string `mapstructure:"policy_file"`
	AutoReload     bool   `mapstructure:"auto_reload"`
	ReloadInterval int    `mapstructure:"reload_interval"` // in seconds
}

// OrchestratorConfig holds orchestrator configuration
type OrchestratorConfig struct {
	Workers int `mapstructure:"workers"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// DSN returns the database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}
