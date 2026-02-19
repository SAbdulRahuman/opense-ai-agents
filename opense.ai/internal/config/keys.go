package config

import "os"

// APIKeySource represents where an API key comes from.
type APIKeySource string

const (
	KeySourceEnv    APIKeySource = "env"
	KeySourceConfig APIKeySource = "config"
	KeySourceNone   APIKeySource = "none"
)

// KeyStatus represents the status of an API key.
type KeyStatus struct {
	Name     string       `json:"name"`
	Source   APIKeySource `json:"source"`
	IsSet    bool         `json:"is_set"`
	Masked   string       `json:"masked,omitempty"` // e.g., "sk-...abc"
}

// CheckAPIKeys returns the status of all required API keys.
func CheckAPIKeys(cfg *Config) []KeyStatus {
	return []KeyStatus{
		checkKey("OpenAI API Key", cfg.LLM.OpenAIKey, "OPENSEAI_LLM_OPENAI_KEY"),
		checkKey("Gemini API Key", cfg.LLM.GeminiKey, "OPENSEAI_LLM_GEMINI_KEY"),
		checkKey("Anthropic API Key", cfg.LLM.AnthropicKey, "OPENSEAI_LLM_ANTHROPIC_KEY"),
		checkKey("Zerodha API Key", cfg.Broker.Zerodha.APIKey, "OPENSEAI_BROKER_ZERODHA_API_KEY"),
		checkKey("Zerodha API Secret", cfg.Broker.Zerodha.APISecret, "OPENSEAI_BROKER_ZERODHA_API_SECRET"),
	}
}

// checkKey checks if a key is set and where it came from.
func checkKey(name, value, envVar string) KeyStatus {
	status := KeyStatus{
		Name:  name,
		IsSet: value != "",
	}

	if value != "" {
		// Check if it came from env
		if os.Getenv(envVar) != "" {
			status.Source = KeySourceEnv
		} else {
			status.Source = KeySourceConfig
		}
		status.Masked = maskKey(value)
	} else {
		status.Source = KeySourceNone
	}

	return status
}

// maskKey masks an API key for display, showing only first 3 and last 3 chars.
func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:3] + "..." + key[len(key)-3:]
}
