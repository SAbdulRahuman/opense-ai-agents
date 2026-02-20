// Package api — configuration management endpoints.
package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/seenimoa/openseai/internal/config"
)

// configMu serialises writes to the config file.
var configMu sync.Mutex

// ConfigResponse is the JSON envelope returned by GET /api/v1/config.
type ConfigResponse struct {
	Config     *config.Config `json:"config"`
	ConfigFile string         `json:"config_file"` // path to the active config file
}

// handleGetConfig returns the current (running) configuration.
// Sensitive keys (API keys/secrets) are excluded via json:"-" tags.
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: ConfigResponse{
			Config:     s.cfg,
			ConfigFile: config.ConfigFilePath(),
		},
	})
}

// handleUpdateConfig merges the provided partial configuration into the running
// config, persists it to disk, and returns the updated config.
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var incoming config.Config
	if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	configMu.Lock()
	defer configMu.Unlock()

	// Merge non-zero values from incoming into running config.
	mergeConfig(s.cfg, &incoming)

	// Persist to disk.
	cfgPath := config.ConfigFilePath()
	if err := config.SaveToFile(s.cfg, cfgPath); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: ConfigResponse{
			Config:     s.cfg,
			ConfigFile: cfgPath,
		},
	})
}

// handleGetConfigKeys returns the status of all sensitive API keys.
func (s *Server) handleGetConfigKeys(w http.ResponseWriter, r *http.Request) {
	keys := config.CheckAPIKeys(s.cfg)
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    keys,
	})
}

// mergeConfig copies non-zero/non-empty values from src into dst.
func mergeConfig(dst, src *config.Config) {
	// LLM
	if src.LLM.Primary != "" {
		dst.LLM.Primary = src.LLM.Primary
	}
	if src.LLM.OllamaURL != "" {
		dst.LLM.OllamaURL = src.LLM.OllamaURL
	}
	if src.LLM.Model != "" {
		dst.LLM.Model = src.LLM.Model
	}
	if src.LLM.FallbackModel != "" {
		dst.LLM.FallbackModel = src.LLM.FallbackModel
	}
	if src.LLM.Temperature != 0 {
		dst.LLM.Temperature = src.LLM.Temperature
	}
	if src.LLM.MaxTokens != 0 {
		dst.LLM.MaxTokens = src.LLM.MaxTokens
	}

	// Broker
	if src.Broker.Provider != "" {
		dst.Broker.Provider = src.Broker.Provider
	}
	if src.Broker.IBKR.Host != "" {
		dst.Broker.IBKR.Host = src.Broker.IBKR.Host
	}
	if src.Broker.IBKR.Port != 0 {
		dst.Broker.IBKR.Port = src.Broker.IBKR.Port
	}

	// Trading
	if src.Trading.Mode != "" {
		dst.Trading.Mode = src.Trading.Mode
	}
	if src.Trading.MaxPositionPct != 0 {
		dst.Trading.MaxPositionPct = src.Trading.MaxPositionPct
	}
	if src.Trading.DailyLossLimitPct != 0 {
		dst.Trading.DailyLossLimitPct = src.Trading.DailyLossLimitPct
	}
	if src.Trading.MaxOpenPositions != 0 {
		dst.Trading.MaxOpenPositions = src.Trading.MaxOpenPositions
	}
	// RequireConfirmation is a bool — always apply from incoming
	dst.Trading.RequireConfirmation = src.Trading.RequireConfirmation
	if src.Trading.ConfirmTimeoutSec != 0 {
		dst.Trading.ConfirmTimeoutSec = src.Trading.ConfirmTimeoutSec
	}
	if src.Trading.InitialCapital != 0 {
		dst.Trading.InitialCapital = src.Trading.InitialCapital
	}

	// Analysis
	if src.Analysis.CacheTTL != 0 {
		dst.Analysis.CacheTTL = src.Analysis.CacheTTL
	}
	if src.Analysis.ConcurrentFetches != 0 {
		dst.Analysis.ConcurrentFetches = src.Analysis.ConcurrentFetches
	}

	// FinanceQL
	if src.FinanceQL.CacheTTL != 0 {
		dst.FinanceQL.CacheTTL = src.FinanceQL.CacheTTL
	}
	if src.FinanceQL.MaxRange != "" {
		dst.FinanceQL.MaxRange = src.FinanceQL.MaxRange
	}
	if src.FinanceQL.AlertCheckInterval != 0 {
		dst.FinanceQL.AlertCheckInterval = src.FinanceQL.AlertCheckInterval
	}
	if src.FinanceQL.REPLHistoryFile != "" {
		dst.FinanceQL.REPLHistoryFile = src.FinanceQL.REPLHistoryFile
	}

	// API
	if src.API.Host != "" {
		dst.API.Host = src.API.Host
	}
	if src.API.Port != 0 {
		dst.API.Port = src.API.Port
	}
	if len(src.API.CORSOrigins) > 0 {
		dst.API.CORSOrigins = src.API.CORSOrigins
	}

	// Web
	if src.Web.URL != "" {
		dst.Web.URL = src.Web.URL
	}

	// Logging
	if src.Logging.Level != "" {
		dst.Logging.Level = src.Logging.Level
	}
	if src.Logging.Format != "" {
		dst.Logging.Format = src.Logging.Format
	}
}
