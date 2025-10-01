package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	HubURL            string        `mapstructure:"hub_url"`
	Token             string        `mapstructure:"token"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	ReconnectWait     time.Duration `mapstructure:"reconnect_wait"`
	OperationTimeout  time.Duration `mapstructure:"operation_timeout"`
	MaxRetries        int           `mapstructure:"max_retries"`
	RetryBackoff      time.Duration `mapstructure:"retry_backoff"`
	Logging           LoggingConfig `mapstructure:"logging"`
}

// LoadAgentConfig loads agent configuration from file and environment variables
func LoadAgentConfig() (*AgentConfig, error) {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/mckmt")

	// Set default values first
	setAgentDefaults()

	// Enable reading from environment variables (highest priority)
	viper.SetEnvPrefix("MCKMT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Use the default config file name
	viper.SetConfigName("agent_config")

	// Read config file (overrides defaults, but env vars override this)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config AgentConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Ensure agent heartbeat interval has a valid default
	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = 30 * time.Second
	}

	return &config, nil
}

// setAgentDefaults sets default values for agent configuration
func setAgentDefaults() {
	viper.SetDefault("hub_url", "localhost:8081")
	viper.SetDefault("token", "")
	viper.SetDefault("heartbeat_interval", "30s")
	viper.SetDefault("reconnect_wait", "5s")
	viper.SetDefault("operation_timeout", "5m")
	viper.SetDefault("max_retries", 3)
	viper.SetDefault("retry_backoff", "1s")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}
