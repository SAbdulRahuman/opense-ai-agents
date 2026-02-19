// OpeNSE.ai — Agentic AI for NSE Stock Analysis & Trading
//
// Main CLI entrypoint using cobra command framework.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/seenimoa/openseai/internal/config"
	"github.com/seenimoa/openseai/pkg/utils"
)

// Build-time variables (set via -ldflags).
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Global config
var cfg *config.Config

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "openseai",
	Short: "OpeNSE.ai — Agentic AI for NSE Stock Analysis & Trading",
	Long: `OpeNSE.ai (Open + NSE + Agentic AI)
A Go-based multi-agent AI system for comprehensive NSE stock analysis,
covering fundamental, technical, derivatives, sentiment analysis, and
automated trading.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		configFile, _ := cmd.Flags().GetString("config")
		if configFile != "" {
			cfg, err = config.LoadFromFile(configFile)
		} else {
			cfg, err = config.Load()
		}
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().String("config", "", "config file path (default: ./config/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "", "log level override (debug, info, warn, error)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(technicalCmd)
	rootCmd.AddCommand(fundamentalCmd)
	rootCmd.AddCommand(fnoCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(statusCmd)
}

// --- Version Command ---

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OpeNSE.ai %s\n", version)
		fmt.Printf("  commit:  %s\n", commit)
		fmt.Printf("  built:   %s\n", date)
	},
}

// --- Analyze Command ---

var analyzeCmd = &cobra.Command{
	Use:   "analyze [ticker]",
	Short: "Run analysis on a stock",
	Long:  "Run single-agent quick analysis or multi-agent deep analysis on a stock.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		deep, _ := cmd.Flags().GetBool("deep")

		mode := "quick (single-agent)"
		if deep {
			mode = "deep (multi-agent)"
		}

		fmt.Printf("🔍 Analyzing %s — %s mode\n", ticker, mode)
		fmt.Printf("   Market Status: %s\n", utils.MarketStatus())
		fmt.Println("\n⚠️  Analysis engine not yet implemented. Coming in Phase 3+5.")
		return nil
	},
}

func init() {
	analyzeCmd.Flags().Bool("deep", false, "run multi-agent deep analysis")
	analyzeCmd.Flags().Bool("pdf", false, "generate PDF report")
}

// --- Technical Command ---

var technicalCmd = &cobra.Command{
	Use:   "technical [ticker]",
	Short: "Run technical analysis on a stock",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		fmt.Printf("📊 Technical Analysis: %s\n", ticker)
		fmt.Println("\n⚠️  Technical analysis engine not yet implemented. Coming in Phase 3.")
		return nil
	},
}

// --- Fundamental Command ---

var fundamentalCmd = &cobra.Command{
	Use:   "fundamental [ticker]",
	Short: "Run fundamental analysis on a stock",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		fmt.Printf("📈 Fundamental Analysis: %s\n", ticker)
		fmt.Println("\n⚠️  Fundamental analysis engine not yet implemented. Coming in Phase 3.")
		return nil
	},
}

// --- F&O Command ---

var fnoCmd = &cobra.Command{
	Use:   "fno [ticker]",
	Short: "Run F&O / derivatives analysis",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		fmt.Printf("🎯 F&O Analysis: %s\n", ticker)
		fmt.Println("\n⚠️  F&O analysis engine not yet implemented. Coming in Phase 3.")
		return nil
	},
}

// --- Query Command (FinanceQL) ---

var queryCmd = &cobra.Command{
	Use:   "query [expression]",
	Short: "Execute a FinanceQL query",
	Long: `Execute a FinanceQL query expression or start the interactive REPL.

Examples:
  openseai query 'rsi(RELIANCE, 14)'
  openseai query 'price(TCS)[30d] | sma(20) | trend()'
  openseai query 'screener(pe < 15 AND roe > 20)'
  openseai query --repl
  openseai query --nl "oversold IT stocks"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repl, _ := cmd.Flags().GetBool("repl")
		nl, _ := cmd.Flags().GetString("nl")

		if repl {
			fmt.Println("📟 FinanceQL Interactive REPL")
			fmt.Println("⚠️  REPL not yet implemented. Coming in Phase 9.")
			return nil
		}

		if nl != "" {
			fmt.Printf("🗣️  Natural Language → FinanceQL: %q\n", nl)
			fmt.Println("⚠️  NL→FinanceQL not yet implemented. Coming in Phase 9.")
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("provide a FinanceQL expression or use --repl")
		}

		fmt.Printf("📟 FinanceQL: %s\n", args[0])
		fmt.Println("⚠️  FinanceQL engine not yet implemented. Coming in Phase 9.")
		return nil
	},
}

func init() {
	queryCmd.Flags().Bool("repl", false, "start interactive FinanceQL REPL")
	queryCmd.Flags().String("nl", "", "natural language query to translate to FinanceQL")
}

// --- Chat Command ---

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("💬 OpeNSE.ai Chat Mode")
		fmt.Println("⚠️  Chat mode not yet implemented. Coming in Phase 5.")
		return nil
	},
}

// --- Serve Command (API Server) ---

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		port := cfg.API.Port
		fmt.Printf("🌐 Starting OpeNSE.ai API server on :%d\n", port)
		fmt.Println("⚠️  API server not yet implemented. Coming in Phase 10.")
		return nil
	},
}

// --- Status Command ---

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status and configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("═══════════════════════════════════════")
		fmt.Println("  OpeNSE.ai — System Status")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("  Version:       %s (%s)\n", version, commit)
		fmt.Printf("  Market Status: %s\n", utils.MarketStatus())
		fmt.Printf("  Time (IST):    %s\n", utils.FormatDateTimeIST(utils.NowIST()))
		fmt.Println()

		// Config summary
		fmt.Println("  Configuration:")
		fmt.Printf("    LLM Provider:  %s (model: %s)\n", cfg.LLM.Primary, cfg.LLM.Model)
		fmt.Printf("    Broker:        %s\n", cfg.Broker.Provider)
		fmt.Printf("    Trading Mode:  %s\n", cfg.Trading.Mode)
		fmt.Printf("    API Server:    %s:%d\n", cfg.API.Host, cfg.API.Port)
		fmt.Println()

		// API keys status
		fmt.Println("  API Keys:")
		keys := config.CheckAPIKeys(cfg)
		for _, k := range keys {
			status := "❌ not set"
			if k.IsSet {
				status = fmt.Sprintf("✅ set (%s: %s)", k.Source, k.Masked)
			}
			fmt.Printf("    %-25s %s\n", k.Name+":", status)
		}

		fmt.Println("═══════════════════════════════════════")
		return nil
	},
}
