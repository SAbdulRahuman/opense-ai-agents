// Package config handles configuration loading for OpeNSE.ai.
// It supports YAML config files with environment variable overrides.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration.
type Config struct {
	LLM        LLMConfig        `mapstructure:"llm"        yaml:"llm"        json:"llm"`
	Broker     BrokerConfig     `mapstructure:"broker"     yaml:"broker"     json:"broker"`
	Trading    TradingConfig    `mapstructure:"trading"    yaml:"trading"    json:"trading"`
	Analysis   AnalysisConfig   `mapstructure:"analysis"   yaml:"analysis"   json:"analysis"`
	FinanceQL  FinanceQLConfig  `mapstructure:"financeql"  yaml:"financeql"  json:"financeql"`
	API        APIConfig        `mapstructure:"api"        yaml:"api"        json:"api"`
	Web        WebConfig        `mapstructure:"web"        yaml:"web"        json:"web"`
	Logging    LoggingConfig    `mapstructure:"logging"    yaml:"logging"    json:"logging"`
}

// LLMConfig holds LLM provider configuration.
type LLMConfig struct {
	Primary      string  `mapstructure:"primary"       yaml:"primary"       json:"primary"`       // "openai", "ollama", "gemini", "anthropic"
	OpenAIKey    string  `mapstructure:"openai_key"     yaml:"openai_key"     json:"-"`             // excluded from JSON — use /config/keys
	OllamaURL    string  `mapstructure:"ollama_url"     yaml:"ollama_url"     json:"ollama_url"`
	GeminiKey    string  `mapstructure:"gemini_key"     yaml:"gemini_key"     json:"-"`
	AnthropicKey string  `mapstructure:"anthropic_key"  yaml:"anthropic_key"  json:"-"`
	Model        string  `mapstructure:"model"          yaml:"model"          json:"model"`
	FallbackModel string `mapstructure:"fallback_model" yaml:"fallback_model" json:"fallback_model"`
	Temperature  float64 `mapstructure:"temperature"   yaml:"temperature"   json:"temperature"`
	MaxTokens    int     `mapstructure:"max_tokens"     yaml:"max_tokens"     json:"max_tokens"`
}

// BrokerConfig holds broker integration configuration.
type BrokerConfig struct {
	Provider string        `mapstructure:"provider" yaml:"provider" json:"provider"` // "paper", "zerodha", "ibkr"
	Zerodha  ZerodhaConfig `mapstructure:"zerodha"  yaml:"zerodha"  json:"zerodha"`
	IBKR     IBKRConfig    `mapstructure:"ibkr"     yaml:"ibkr"     json:"ibkr"`
}

// ZerodhaConfig holds Zerodha Kite API credentials.
type ZerodhaConfig struct {
	APIKey    string `mapstructure:"api_key"    yaml:"api_key"    json:"-"`
	APISecret string `mapstructure:"api_secret" yaml:"api_secret" json:"-"`
}

// IBKRConfig holds Interactive Brokers connection settings.
type IBKRConfig struct {
	Host string `mapstructure:"host" yaml:"host" json:"host"`
	Port int    `mapstructure:"port" yaml:"port" json:"port"`
}

// TradingConfig holds trading safety and risk management settings.
type TradingConfig struct {
	Mode                string  `mapstructure:"mode"                  yaml:"mode"                  json:"mode"`
	MaxPositionPct      float64 `mapstructure:"max_position_pct"      yaml:"max_position_pct"      json:"max_position_pct"`
	DailyLossLimitPct   float64 `mapstructure:"daily_loss_limit_pct"  yaml:"daily_loss_limit_pct"  json:"daily_loss_limit_pct"`
	MaxOpenPositions    int     `mapstructure:"max_open_positions"    yaml:"max_open_positions"    json:"max_open_positions"`
	RequireConfirmation bool    `mapstructure:"require_confirmation"  yaml:"require_confirmation"  json:"require_confirmation"`
	ConfirmTimeoutSec   int     `mapstructure:"confirm_timeout_sec"   yaml:"confirm_timeout_sec"   json:"confirm_timeout_sec"`
	InitialCapital      float64 `mapstructure:"initial_capital"       yaml:"initial_capital"       json:"initial_capital"`
}

// AnalysisConfig holds analysis engine settings.
type AnalysisConfig struct {
	CacheTTL         int `mapstructure:"cache_ttl"          yaml:"cache_ttl"          json:"cache_ttl"`
	ConcurrentFetches int `mapstructure:"concurrent_fetches" yaml:"concurrent_fetches" json:"concurrent_fetches"`
}

// FinanceQLConfig holds FinanceQL query language settings.
type FinanceQLConfig struct {
	CacheTTL            int    `mapstructure:"cache_ttl"              yaml:"cache_ttl"              json:"cache_ttl"`
	MaxRange            string `mapstructure:"max_range"              yaml:"max_range"              json:"max_range"`
	AlertCheckInterval  int    `mapstructure:"alert_check_interval"   yaml:"alert_check_interval"   json:"alert_check_interval"`
	REPLHistoryFile     string `mapstructure:"repl_history_file"      yaml:"repl_history_file"      json:"repl_history_file"`
}

// APIConfig holds HTTP/gRPC API server settings.
type APIConfig struct {
	Host        string   `mapstructure:"host"         yaml:"host"         json:"host"`
	Port        int      `mapstructure:"port"         yaml:"port"         json:"port"`
	CORSOrigins []string `mapstructure:"cors_origins"  yaml:"cors_origins"  json:"cors_origins"`
}

// WebConfig holds Next.js frontend configuration.
type WebConfig struct {
	URL string `mapstructure:"url" yaml:"url" json:"url"` // e.g., "http://localhost:3000"
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"  yaml:"level"  json:"level"`  // "debug", "info", "warn", "error"
	Format string `mapstructure:"format" yaml:"format" json:"format"` // "text" or "json"
}

// Load reads the configuration from file and environment variables.
// Config file search order:
//  1. ./config/config.yaml (project root)
//  2. ~/.openseai/config.yaml (home directory)
//  3. /etc/openseai/config.yaml (system)
//
// Environment variables override config file values.
// Format: OPENSEAI_<SECTION>_<KEY>, e.g., OPENSEAI_LLM_OPENAI_KEY
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file settings
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(filepath.Join(homeDir(), ".openseai"))
	v.AddConfigPath("/etc/openseai")

	// Environment variable settings
	v.SetEnvPrefix("OPENSEAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file (not required to exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found — that's fine, use defaults + env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override sensitive values from environment
	overrideFromEnv(&cfg)

	return &cfg, nil
}

// LoadFromFile reads configuration from a specific file path.
func LoadFromFile(path string) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	v.SetConfigFile(path)
	v.SetEnvPrefix("OPENSEAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	overrideFromEnv(&cfg)
	return &cfg, nil
}

// setDefaults sets sensible defaults for all config values.
func setDefaults(v *viper.Viper) {
	// LLM defaults
	v.SetDefault("llm.primary", "openai")
	v.SetDefault("llm.ollama_url", "http://localhost:11434")
	v.SetDefault("llm.model", "gpt-4o")
	v.SetDefault("llm.temperature", 0.1)
	v.SetDefault("llm.max_tokens", 4096)

	// Broker defaults
	v.SetDefault("broker.provider", "paper")
	v.SetDefault("broker.ibkr.host", "127.0.0.1")
	v.SetDefault("broker.ibkr.port", 7497)

	// Trading defaults (safety-first)
	v.SetDefault("trading.mode", "paper")
	v.SetDefault("trading.max_position_pct", 5.0)
	v.SetDefault("trading.daily_loss_limit_pct", 2.0)
	v.SetDefault("trading.max_open_positions", 10)
	v.SetDefault("trading.require_confirmation", true)
	v.SetDefault("trading.confirm_timeout_sec", 60)
	v.SetDefault("trading.initial_capital", 1000000) // ₹10 lakh default

	// Analysis defaults
	v.SetDefault("analysis.cache_ttl", 300)          // 5 minutes
	v.SetDefault("analysis.concurrent_fetches", 5)

	// FinanceQL defaults
	v.SetDefault("financeql.cache_ttl", 60)           // 1 minute
	v.SetDefault("financeql.max_range", "365d")
	v.SetDefault("financeql.alert_check_interval", 30)
	v.SetDefault("financeql.repl_history_file", "~/.openseai/financeql_history")

	// API defaults
	v.SetDefault("api.host", "0.0.0.0")
	v.SetDefault("api.port", 8080)
	v.SetDefault("api.cors_origins", []string{"http://localhost:3000"})

	// Web defaults
	v.SetDefault("web.url", "http://localhost:3000")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
}

// overrideFromEnv explicitly reads sensitive keys from environment variables.
func overrideFromEnv(cfg *Config) {
	if key := os.Getenv("OPENSEAI_LLM_OPENAI_KEY"); key != "" {
		cfg.LLM.OpenAIKey = key
	}
	if key := os.Getenv("OPENSEAI_LLM_GEMINI_KEY"); key != "" {
		cfg.LLM.GeminiKey = key
	}
	if key := os.Getenv("OPENSEAI_LLM_ANTHROPIC_KEY"); key != "" {
		cfg.LLM.AnthropicKey = key
	}
	if key := os.Getenv("OPENSEAI_BROKER_ZERODHA_API_KEY"); key != "" {
		cfg.Broker.Zerodha.APIKey = key
	}
	if key := os.Getenv("OPENSEAI_BROKER_ZERODHA_API_SECRET"); key != "" {
		cfg.Broker.Zerodha.APISecret = key
	}
}

// SaveToFile writes the current configuration to a YAML file.
// If path is empty, it writes to ./config/config.yaml.
func SaveToFile(cfg *Config, path string) error {
	if path == "" {
		path = filepath.Join(".", "config", "config.yaml")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", path, err)
	}
	return nil
}

// ConfigFilePath returns the path to the active config file (if any).
// Returns empty string if no config file was found.
func ConfigFilePath() string {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(filepath.Join(homeDir(), ".openseai"))
	v.AddConfigPath("/etc/openseai")

	if err := v.ReadInConfig(); err != nil {
		return ""
	}
	return v.ConfigFileUsed()
}

// homeDir returns the user's home directory.
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
