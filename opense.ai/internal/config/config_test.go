package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ── Load / Defaults ──

func TestLoadReturnsDefaults(t *testing.T) {
	// Unset any env vars that would interfere
	envVars := []string{
		"OPENSEAI_LLM_OPENAI_KEY", "OPENSEAI_LLM_GEMINI_KEY", "OPENSEAI_LLM_ANTHROPIC_KEY",
		"OPENSEAI_BROKER_ZERODHA_API_KEY", "OPENSEAI_BROKER_ZERODHA_API_SECRET",
	}
	for _, e := range envVars {
		os.Unsetenv(e)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// LLM defaults
	if cfg.LLM.Primary != "openai" {
		t.Errorf("LLM.Primary: got %q, want %q", cfg.LLM.Primary, "openai")
	}
	if cfg.LLM.Model != "gpt-4o" {
		t.Errorf("LLM.Model: got %q, want %q", cfg.LLM.Model, "gpt-4o")
	}
	if cfg.LLM.Temperature != 0.1 {
		t.Errorf("LLM.Temperature: got %f, want 0.1", cfg.LLM.Temperature)
	}
	if cfg.LLM.MaxTokens != 4096 {
		t.Errorf("LLM.MaxTokens: got %d, want 4096", cfg.LLM.MaxTokens)
	}
	if cfg.LLM.OllamaURL != "http://localhost:11434" {
		t.Errorf("LLM.OllamaURL: got %q", cfg.LLM.OllamaURL)
	}

	// Broker defaults
	if cfg.Broker.Provider != "paper" {
		t.Errorf("Broker.Provider: got %q, want %q", cfg.Broker.Provider, "paper")
	}
	if cfg.Broker.IBKR.Host != "127.0.0.1" {
		t.Errorf("Broker.IBKR.Host: got %q, want %q", cfg.Broker.IBKR.Host, "127.0.0.1")
	}
	if cfg.Broker.IBKR.Port != 7497 {
		t.Errorf("Broker.IBKR.Port: got %d, want 7497", cfg.Broker.IBKR.Port)
	}

	// Trading defaults
	if cfg.Trading.Mode != "paper" {
		t.Errorf("Trading.Mode: got %q, want %q", cfg.Trading.Mode, "paper")
	}
	if cfg.Trading.MaxPositionPct != 5.0 {
		t.Errorf("Trading.MaxPositionPct: got %f, want 5.0", cfg.Trading.MaxPositionPct)
	}
	if cfg.Trading.DailyLossLimitPct != 2.0 {
		t.Errorf("Trading.DailyLossLimitPct: got %f, want 2.0", cfg.Trading.DailyLossLimitPct)
	}
	if cfg.Trading.MaxOpenPositions != 10 {
		t.Errorf("Trading.MaxOpenPositions: got %d, want 10", cfg.Trading.MaxOpenPositions)
	}
	if !cfg.Trading.RequireConfirmation {
		t.Error("Trading.RequireConfirmation should be true by default")
	}
	if cfg.Trading.ConfirmTimeoutSec != 60 {
		t.Errorf("Trading.ConfirmTimeoutSec: got %d, want 60", cfg.Trading.ConfirmTimeoutSec)
	}
	if cfg.Trading.InitialCapital != 1000000 {
		t.Errorf("Trading.InitialCapital: got %f, want 1000000", cfg.Trading.InitialCapital)
	}

	// Analysis defaults
	if cfg.Analysis.CacheTTL != 300 {
		t.Errorf("Analysis.CacheTTL: got %d, want 300", cfg.Analysis.CacheTTL)
	}
	if cfg.Analysis.ConcurrentFetches != 5 {
		t.Errorf("Analysis.ConcurrentFetches: got %d, want 5", cfg.Analysis.ConcurrentFetches)
	}

	// FinanceQL defaults
	if cfg.FinanceQL.CacheTTL != 60 {
		t.Errorf("FinanceQL.CacheTTL: got %d, want 60", cfg.FinanceQL.CacheTTL)
	}
	if cfg.FinanceQL.MaxRange != "365d" {
		t.Errorf("FinanceQL.MaxRange: got %q, want %q", cfg.FinanceQL.MaxRange, "365d")
	}
	if cfg.FinanceQL.AlertCheckInterval != 30 {
		t.Errorf("FinanceQL.AlertCheckInterval: got %d, want 30", cfg.FinanceQL.AlertCheckInterval)
	}

	// API defaults
	if cfg.API.Host != "0.0.0.0" {
		t.Errorf("API.Host: got %q, want %q", cfg.API.Host, "0.0.0.0")
	}
	if cfg.API.Port != 8080 {
		t.Errorf("API.Port: got %d, want 8080", cfg.API.Port)
	}

	// Web defaults
	if cfg.Web.URL != "http://localhost:3000" {
		t.Errorf("Web.URL: got %q", cfg.Web.URL)
	}

	// Logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level: got %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format: got %q, want %q", cfg.Logging.Format, "text")
	}
}

// ── LoadFromFile ──

func TestLoadFromFile(t *testing.T) {
	// Create a temp config file
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test_config.yaml")
	content := []byte(`
llm:
  primary: "gemini"
  model: "gemini-pro"
  temperature: 0.3
  max_tokens: 8192
broker:
  provider: "zerodha"
  zerodha:
    api_key: "test_key_12345678901234"
    api_secret: "test_secret_1234567890"
trading:
  mode: "paper"
  max_position_pct: 3.0
  initial_capital: 2000000
api:
  port: 9090
logging:
  level: "debug"
  format: "json"
`)
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	// Unset env vars
	os.Unsetenv("OPENSEAI_LLM_OPENAI_KEY")
	os.Unsetenv("OPENSEAI_BROKER_ZERODHA_API_KEY")
	os.Unsetenv("OPENSEAI_BROKER_ZERODHA_API_SECRET")

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("LoadFromFile() error: %v", err)
	}
	if cfg.LLM.Primary != "gemini" {
		t.Errorf("LLM.Primary: got %q, want %q", cfg.LLM.Primary, "gemini")
	}
	if cfg.LLM.Model != "gemini-pro" {
		t.Errorf("LLM.Model: got %q, want %q", cfg.LLM.Model, "gemini-pro")
	}
	if cfg.LLM.Temperature != 0.3 {
		t.Errorf("LLM.Temperature: got %f, want 0.3", cfg.LLM.Temperature)
	}
	if cfg.LLM.MaxTokens != 8192 {
		t.Errorf("LLM.MaxTokens: got %d, want 8192", cfg.LLM.MaxTokens)
	}
	if cfg.Broker.Provider != "zerodha" {
		t.Errorf("Broker.Provider: got %q, want %q", cfg.Broker.Provider, "zerodha")
	}
	if cfg.Broker.Zerodha.APIKey != "test_key_12345678901234" {
		t.Errorf("Broker.Zerodha.APIKey: got %q", cfg.Broker.Zerodha.APIKey)
	}
	if cfg.Trading.MaxPositionPct != 3.0 {
		t.Errorf("Trading.MaxPositionPct: got %f, want 3.0", cfg.Trading.MaxPositionPct)
	}
	if cfg.Trading.InitialCapital != 2000000 {
		t.Errorf("Trading.InitialCapital: got %f, want 2000000", cfg.Trading.InitialCapital)
	}
	if cfg.API.Port != 9090 {
		t.Errorf("API.Port: got %d, want 9090", cfg.API.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level: got %q, want %q", cfg.Logging.Level, "debug")
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format: got %q, want %q", cfg.Logging.Format, "json")
	}
}

func TestLoadFromFileNotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("LoadFromFile() with nonexistent path should return error")
	}
}

// ── overrideFromEnv ──

func TestOverrideFromEnv(t *testing.T) {
	cfg := &Config{}

	// Set env vars
	os.Setenv("OPENSEAI_LLM_OPENAI_KEY", "sk-test-openai-key-123456")
	os.Setenv("OPENSEAI_LLM_GEMINI_KEY", "gemini-key-789")
	os.Setenv("OPENSEAI_LLM_ANTHROPIC_KEY", "sk-ant-test")
	os.Setenv("OPENSEAI_BROKER_ZERODHA_API_KEY", "zerodha-api-key")
	os.Setenv("OPENSEAI_BROKER_ZERODHA_API_SECRET", "zerodha-secret")
	defer func() {
		os.Unsetenv("OPENSEAI_LLM_OPENAI_KEY")
		os.Unsetenv("OPENSEAI_LLM_GEMINI_KEY")
		os.Unsetenv("OPENSEAI_LLM_ANTHROPIC_KEY")
		os.Unsetenv("OPENSEAI_BROKER_ZERODHA_API_KEY")
		os.Unsetenv("OPENSEAI_BROKER_ZERODHA_API_SECRET")
	}()

	overrideFromEnv(cfg)

	if cfg.LLM.OpenAIKey != "sk-test-openai-key-123456" {
		t.Errorf("OpenAIKey: got %q", cfg.LLM.OpenAIKey)
	}
	if cfg.LLM.GeminiKey != "gemini-key-789" {
		t.Errorf("GeminiKey: got %q", cfg.LLM.GeminiKey)
	}
	if cfg.LLM.AnthropicKey != "sk-ant-test" {
		t.Errorf("AnthropicKey: got %q", cfg.LLM.AnthropicKey)
	}
	if cfg.Broker.Zerodha.APIKey != "zerodha-api-key" {
		t.Errorf("Zerodha.APIKey: got %q", cfg.Broker.Zerodha.APIKey)
	}
	if cfg.Broker.Zerodha.APISecret != "zerodha-secret" {
		t.Errorf("Zerodha.APISecret: got %q", cfg.Broker.Zerodha.APISecret)
	}
}

func TestOverrideFromEnvNoEnvSet(t *testing.T) {
	os.Unsetenv("OPENSEAI_LLM_OPENAI_KEY")
	os.Unsetenv("OPENSEAI_LLM_GEMINI_KEY")
	os.Unsetenv("OPENSEAI_LLM_ANTHROPIC_KEY")
	os.Unsetenv("OPENSEAI_BROKER_ZERODHA_API_KEY")
	os.Unsetenv("OPENSEAI_BROKER_ZERODHA_API_SECRET")

	cfg := &Config{
		LLM: LLMConfig{OpenAIKey: "from-config"},
	}
	overrideFromEnv(cfg)

	// Should retain the original value when env is not set
	if cfg.LLM.OpenAIKey != "from-config" {
		t.Errorf("OpenAIKey should stay as 'from-config' when env is unset, got %q", cfg.LLM.OpenAIKey)
	}
}

// ── maskKey ──

func TestMaskKeyShort(t *testing.T) {
	// Keys with 8 or fewer characters should be fully masked
	tests := []struct {
		input string
		want  string
	}{
		{"", "***"},
		{"a", "***"},
		{"abcd", "***"},
		{"12345678", "***"},
	}
	for _, tc := range tests {
		got := maskKey(tc.input)
		if got != tc.want {
			t.Errorf("maskKey(%q): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestMaskKeyLong(t *testing.T) {
	// Keys with more than 8 characters show first 3 + ... + last 3
	tests := []struct {
		input string
		want  string
	}{
		{"123456789", "123...789"},
		{"sk-abcdef1234567890xyz", "sk-...xyz"},
		{"ABCDEFGHIJKLMNOP", "ABC...NOP"},
	}
	for _, tc := range tests {
		got := maskKey(tc.input)
		if got != tc.want {
			t.Errorf("maskKey(%q): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── CheckAPIKeys / checkKey ──

func TestCheckAPIKeysAllEmpty(t *testing.T) {
	// Clear env vars
	envVars := []string{
		"OPENSEAI_LLM_OPENAI_KEY", "OPENSEAI_LLM_GEMINI_KEY", "OPENSEAI_LLM_ANTHROPIC_KEY",
		"OPENSEAI_BROKER_ZERODHA_API_KEY", "OPENSEAI_BROKER_ZERODHA_API_SECRET",
	}
	for _, e := range envVars {
		os.Unsetenv(e)
	}

	cfg := &Config{}
	statuses := CheckAPIKeys(cfg)

	if len(statuses) != 5 {
		t.Fatalf("CheckAPIKeys: got %d statuses, want 5", len(statuses))
	}
	for _, s := range statuses {
		if s.IsSet {
			t.Errorf("Key %q should not be set", s.Name)
		}
		if s.Source != KeySourceNone {
			t.Errorf("Key %q source: got %q, want %q", s.Name, s.Source, KeySourceNone)
		}
	}
}

func TestCheckAPIKeysFromConfig(t *testing.T) {
	os.Unsetenv("OPENSEAI_LLM_OPENAI_KEY")

	cfg := &Config{
		LLM: LLMConfig{
			OpenAIKey: "sk-test-very-long-key-value",
		},
	}
	statuses := CheckAPIKeys(cfg)

	found := false
	for _, s := range statuses {
		if s.Name == "OpenAI API Key" {
			found = true
			if !s.IsSet {
				t.Error("OpenAI key should be set")
			}
			if s.Source != KeySourceConfig {
				t.Errorf("Source: got %q, want %q", s.Source, KeySourceConfig)
			}
			if s.Masked != "sk-...lue" {
				t.Errorf("Masked: got %q, want %q", s.Masked, "sk-...lue")
			}
		}
	}
	if !found {
		t.Error("OpenAI API Key status not found")
	}
}

func TestCheckAPIKeysFromEnv(t *testing.T) {
	os.Setenv("OPENSEAI_LLM_OPENAI_KEY", "sk-env-key-for-testing")
	defer os.Unsetenv("OPENSEAI_LLM_OPENAI_KEY")

	cfg := &Config{
		LLM: LLMConfig{
			OpenAIKey: "sk-env-key-for-testing",
		},
	}
	statuses := CheckAPIKeys(cfg)

	for _, s := range statuses {
		if s.Name == "OpenAI API Key" {
			if s.Source != KeySourceEnv {
				t.Errorf("Source: got %q, want %q", s.Source, KeySourceEnv)
			}
		}
	}
}

func TestCheckKeySourceDetection(t *testing.T) {
	// No env, no value
	os.Unsetenv("TEST_VAR")
	s := checkKey("Test", "", "TEST_VAR")
	if s.Source != KeySourceNone {
		t.Errorf("empty value: got source %q, want %q", s.Source, KeySourceNone)
	}
	if s.IsSet {
		t.Error("empty value should not be set")
	}

	// Value from config (no env)
	s = checkKey("Test", "config-value-long-enough", "TEST_VAR")
	if s.Source != KeySourceConfig {
		t.Errorf("config value: got source %q, want %q", s.Source, KeySourceConfig)
	}
	if !s.IsSet {
		t.Error("config value should be set")
	}

	// Value from env
	os.Setenv("TEST_VAR", "env-value-long-enough")
	defer os.Unsetenv("TEST_VAR")
	s = checkKey("Test", "env-value-long-enough", "TEST_VAR")
	if s.Source != KeySourceEnv {
		t.Errorf("env value: got source %q, want %q", s.Source, KeySourceEnv)
	}
}

// ── homeDir ──

func TestHomeDirReturnsNonEmpty(t *testing.T) {
	h := homeDir()
	if h == "" {
		t.Error("homeDir() should not return empty string")
	}
}

// ── APIKeySource constants ──

func TestAPIKeySourceConstants(t *testing.T) {
	if string(KeySourceEnv) != "env" {
		t.Errorf("KeySourceEnv: got %q", KeySourceEnv)
	}
	if string(KeySourceConfig) != "config" {
		t.Errorf("KeySourceConfig: got %q", KeySourceConfig)
	}
	if string(KeySourceNone) != "none" {
		t.Errorf("KeySourceNone: got %q", KeySourceNone)
	}
}
