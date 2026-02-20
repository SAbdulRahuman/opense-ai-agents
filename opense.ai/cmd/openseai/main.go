// OpeNSE.ai — Agentic AI for NSE Stock Analysis & Trading
//
// Main CLI entrypoint using cobra command framework.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/seenimoa/openseai/api"
	"github.com/seenimoa/openseai/internal/agent"
	"github.com/seenimoa/openseai/internal/backtest"
	"github.com/seenimoa/openseai/internal/broker"
	"github.com/seenimoa/openseai/internal/config"
	"github.com/seenimoa/openseai/internal/datasource"
	"github.com/seenimoa/openseai/internal/financeql"
	"github.com/seenimoa/openseai/internal/llm"
	"github.com/seenimoa/openseai/internal/report"
	"github.com/seenimoa/openseai/pkg/models"
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
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(backtestCmd)
	rootCmd.AddCommand(tradeCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(portfolioCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(statusCmd)
}

// --- Helper: create orchestrator ---

func newOrchestrator() (*agent.Orchestrator, error) {
	agg := datasource.NewAggregator()
	router, err := llm.NewRouterFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("LLM setup failed: %w", err)
	}
	opts := &llm.ChatOptions{
		Model:       cfg.LLM.Model,
		Temperature: cfg.LLM.Temperature,
		MaxTokens:   cfg.LLM.MaxTokens,
	}
	orch := agent.NewOrchestrator(agent.OrchestratorConfig{
		Provider:    router,
		Aggregator:  agg,
		ChatOptions: opts,
		DefaultMode: agent.ModeSingle,
		Capital:     cfg.Trading.InitialCapital,
	})
	return orch, nil
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
		outputJSON, _ := cmd.Flags().GetBool("json")

		mode := "quick (single-agent)"
		if deep {
			mode = "deep (multi-agent)"
		}

		fmt.Printf("🔍 Analyzing %s — %s mode\n", ticker, mode)
		fmt.Printf("   Market Status: %s\n", utils.MarketStatus())
		fmt.Println()

		orch, err := newOrchestrator()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		var result *agent.AgentResult
		if deep {
			result, err = orch.FullAnalysis(ctx, ticker)
		} else {
			result, err = orch.QuickQuery(ctx, fmt.Sprintf("Analyze %s stock", ticker))
		}
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		printAgentResult(result)
		return nil
	},
}

func init() {
	analyzeCmd.Flags().Bool("deep", false, "run multi-agent deep analysis")
	analyzeCmd.Flags().Bool("json", false, "output result as JSON")
	analyzeCmd.Flags().Bool("pdf", false, "generate PDF report after analysis")
}

// --- Technical Command ---

var technicalCmd = &cobra.Command{
	Use:   "technical [ticker]",
	Short: "Run technical analysis on a stock",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		outputJSON, _ := cmd.Flags().GetBool("json")

		fmt.Printf("📊 Technical Analysis: %s\n", ticker)
		fmt.Println()

		orch, err := newOrchestrator()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := orch.QuickQuery(ctx, fmt.Sprintf("Run technical analysis on %s", ticker))
		if err != nil {
			return fmt.Errorf("technical analysis failed: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		printAgentResult(result)
		return nil
	},
}

func init() {
	technicalCmd.Flags().Bool("json", false, "output result as JSON")
}

// --- Fundamental Command ---

var fundamentalCmd = &cobra.Command{
	Use:   "fundamental [ticker]",
	Short: "Run fundamental analysis on a stock",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		outputJSON, _ := cmd.Flags().GetBool("json")

		fmt.Printf("📈 Fundamental Analysis: %s\n", ticker)
		fmt.Println()

		orch, err := newOrchestrator()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := orch.QuickQuery(ctx, fmt.Sprintf("Run fundamental analysis on %s", ticker))
		if err != nil {
			return fmt.Errorf("fundamental analysis failed: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		printAgentResult(result)
		return nil
	},
}

func init() {
	fundamentalCmd.Flags().Bool("json", false, "output result as JSON")
}

// --- F&O Command ---

var fnoCmd = &cobra.Command{
	Use:   "fno [ticker]",
	Short: "Run F&O / derivatives analysis",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		outputJSON, _ := cmd.Flags().GetBool("json")

		fmt.Printf("🎯 F&O Analysis: %s\n", ticker)
		fmt.Println()

		orch, err := newOrchestrator()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		result, err := orch.QuickQuery(ctx, fmt.Sprintf("Run F&O derivatives analysis on %s", ticker))
		if err != nil {
			return fmt.Errorf("F&O analysis failed: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		printAgentResult(result)
		return nil
	},
}

func init() {
	fnoCmd.Flags().Bool("json", false, "output result as JSON")
}

// --- Report Command ---

var reportCmd = &cobra.Command{
	Use:   "report [ticker]",
	Short: "Generate a research report for a stock",
	Long:  "Run multi-agent deep analysis and generate an HTML or PDF research report.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ticker := utils.NormalizeTicker(args[0])
		pdfFlag, _ := cmd.Flags().GetBool("pdf")
		output, _ := cmd.Flags().GetString("output")

		fmt.Printf("📝 Generating report for %s\n", ticker)
		fmt.Println()

		orch, err := newOrchestrator()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Run deep analysis
		result, err := orch.FullAnalysis(ctx, ticker)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		// Build composite analysis from result
		composite := buildCompositeAnalysis(ticker, result)

		// Generate HTML report
		reportCfg := report.DefaultReportConfig()
		reportCfg.Title = fmt.Sprintf("OpeNSE.ai Research Report — %s", ticker)
		reportCfg.Author = "OpeNSE.ai"
		reportCfg.Sections = report.AllSections()

		html, err := report.GenerateHTML(composite, reportCfg)
		if err != nil {
			return fmt.Errorf("report generation failed: %w", err)
		}

		if pdfFlag {
			if !report.IsPDFSupported() {
				fmt.Println("⚠️  PDF engine not available. Install wkhtmltopdf or chromium.")
				fmt.Println("   Falling back to HTML output.")
				pdfFlag = false
			}
		}

		if pdfFlag {
			if output == "" {
				output = fmt.Sprintf("%s_report_%s.pdf", ticker, time.Now().Format("20060102"))
			}
			pdfCfg := report.DefaultPDFConfig()
			pdfCfg.OutputPath = output
			if err := report.GeneratePDF(html, pdfCfg); err != nil {
				return fmt.Errorf("PDF generation failed: %w", err)
			}
			fmt.Printf("✅ PDF report saved: %s\n", output)
		} else {
			if output == "" {
				output = fmt.Sprintf("%s_report_%s.html", ticker, time.Now().Format("20060102"))
			}
			if err := os.WriteFile(output, []byte(html), 0644); err != nil {
				return fmt.Errorf("failed to write HTML report: %w", err)
			}
			fmt.Printf("✅ HTML report saved: %s\n", output)
		}

		return nil
	},
}

func init() {
	reportCmd.Flags().Bool("pdf", false, "generate PDF report (requires wkhtmltopdf or chromium)")
	reportCmd.Flags().StringP("output", "o", "", "output file path")
}

// --- Backtest Command ---

var backtestCmd = &cobra.Command{
	Use:   "backtest",
	Short: "Run a backtesting simulation",
	Long: `Run a backtest with a built-in or custom strategy.

Available strategies: sma_crossover, rsi_mean_reversion, supertrend, vwap_breakout, macd_crossover

Examples:
  openseai backtest --strategy sma_crossover --ticker RELIANCE --from 2023-01-01
  openseai backtest --strategy rsi_mean_reversion --ticker TCS --from 2024-01-01 --capital 500000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		strategyName, _ := cmd.Flags().GetString("strategy")
		ticker, _ := cmd.Flags().GetString("ticker")
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		capital, _ := cmd.Flags().GetFloat64("capital")
		outputJSON, _ := cmd.Flags().GetBool("json")

		if strategyName == "" || ticker == "" {
			return fmt.Errorf("--strategy and --ticker are required")
		}

		ticker = utils.NormalizeTicker(ticker)

		// Parse dates
		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		to := time.Now()
		if toStr != "" {
			to, err = time.Parse("2006-01-02", toStr)
			if err != nil {
				return fmt.Errorf("invalid --to date: %w", err)
			}
		}

		fmt.Printf("📉 Backtesting %s on %s (%s to %s)\n", strategyName, ticker,
			from.Format("2006-01-02"), to.Format("2006-01-02"))
		fmt.Println()

		// Find strategy
		strategy := findStrategy(strategyName)
		if strategy == nil {
			available := listStrategyNames()
			return fmt.Errorf("unknown strategy %q; available: %s", strategyName, strings.Join(available, ", "))
		}

		// Fetch historical data
		agg := datasource.NewAggregator()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		bars, err := agg.FetchHistoricalData(ctx, ticker, from, to, models.Timeframe1Day)
		if err != nil {
			return fmt.Errorf("failed to fetch data: %w", err)
		}

		if len(bars) < 50 {
			return fmt.Errorf("insufficient data: got %d bars, need at least 50", len(bars))
		}

		// Configure and run
		btCfg := backtest.DefaultConfig()
		if capital > 0 {
			btCfg.InitialCapital = capital
		} else if cfg.Trading.InitialCapital > 0 {
			btCfg.InitialCapital = cfg.Trading.InitialCapital
		}

		engine := backtest.NewEngine(btCfg)
		result, err := engine.Run(strategy, ticker, bars)
		if err != nil {
			return fmt.Errorf("backtest failed: %w", err)
		}

		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		printBacktestResult(result)
		return nil
	},
}

func init() {
	backtestCmd.Flags().StringP("strategy", "s", "", "strategy name (required)")
	backtestCmd.Flags().StringP("ticker", "t", "", "ticker symbol (required)")
	backtestCmd.Flags().String("from", "2023-01-01", "start date (YYYY-MM-DD)")
	backtestCmd.Flags().String("to", "", "end date (YYYY-MM-DD, default: today)")
	backtestCmd.Flags().Float64("capital", 0, "initial capital (default from config)")
	backtestCmd.Flags().Bool("json", false, "output result as JSON")
}

// --- Trade Command ---

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Interactive trading mode",
	Long: `Enter interactive trading mode with paper or live broker.

The trade command provides a REPL-style interface for placing and managing orders
with built-in risk management and human-in-the-loop confirmation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔔 OpeNSE.ai — Interactive Trading Mode")
		fmt.Printf("   Broker: %s\n", cfg.Broker.Provider)
		fmt.Printf("   Mode:   %s\n", cfg.Trading.Mode)
		fmt.Println()

		b := broker.NewPaperBroker(nil)
		riskCfg := broker.DefaultRiskConfig()
		riskCfg.MaxPositionPct = cfg.Trading.MaxPositionPct
		riskCfg.DailyLossLimitPct = cfg.Trading.DailyLossLimitPct
		riskCfg.MaxOpenPositions = cfg.Trading.MaxOpenPositions
		riskCfg.RequireApproval = cfg.Trading.RequireConfirmation
		rm := broker.NewRiskManager(b, riskCfg)

		// Show current portfolio
		ctx := context.Background()
		margins, err := rm.GetMargins(ctx)
		if err == nil {
			fmt.Printf("   Capital:    %s\n", utils.FormatINR(margins.AvailableCash))
			fmt.Println()
		}

		fmt.Println("Commands: buy, sell, positions, orders, margins, cancel, quit")
		fmt.Println("Example: buy RELIANCE 10 2850.00")
		fmt.Println()

		return runTradeREPL(ctx, rm)
	},
}

// --- Watch Command ---

var watchCmd = &cobra.Command{
	Use:   "watch [tickers...]",
	Short: "Real-time watchlist with alerts",
	Long:  "Monitor stocks in real-time with price updates and alert triggers.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, _ := cmd.Flags().GetInt("interval")

		tickers := make([]string, len(args))
		for i, t := range args {
			tickers[i] = utils.NormalizeTicker(t)
		}

		fmt.Printf("👀 Watching: %s (refresh every %ds)\n", strings.Join(tickers, ", "), interval)
		fmt.Println("   Press Ctrl+C to stop")
		fmt.Println()

		agg := datasource.NewAggregator()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle interrupt
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		tickerTimer := time.NewTicker(time.Duration(interval) * time.Second)
		defer tickerTimer.Stop()

		// Initial fetch
		printWatchlist(ctx, agg, tickers)

		for {
			select {
			case <-ctx.Done():
				fmt.Println("\n👋 Stopped watching.")
				return nil
			case <-tickerTimer.C:
				printWatchlist(ctx, agg, tickers)
			}
		}
	},
}

func init() {
	watchCmd.Flags().Int("interval", 30, "refresh interval in seconds")
}

// --- Portfolio Command ---

var portfolioCmd = &cobra.Command{
	Use:   "portfolio",
	Short: "Portfolio analysis from broker",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputJSON, _ := cmd.Flags().GetBool("json")

		fmt.Println("💼 Portfolio Summary")
		fmt.Println()

		b := broker.NewPaperBroker(nil)
		ctx := context.Background()

		margins, err := b.GetMargins(ctx)
		if err != nil {
			return fmt.Errorf("failed to get margins: %w", err)
		}

		positions, err := b.GetPositions(ctx)
		if err != nil {
			return fmt.Errorf("failed to get positions: %w", err)
		}

		holdings, err := b.GetHoldings(ctx)
		if err != nil {
			return fmt.Errorf("failed to get holdings: %w", err)
		}

		orders, err := b.GetOrders(ctx)
		if err != nil {
			return fmt.Errorf("failed to get orders: %w", err)
		}

		if outputJSON {
			data := map[string]any{
				"margins":   margins,
				"positions": positions,
				"holdings":  holdings,
				"orders":    orders,
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		}

		// Print margins
		fmt.Println("═══ Margins ═══")
		fmt.Printf("  Available:  %s\n", utils.FormatINR(margins.AvailableCash))
		fmt.Printf("  Used:       %s\n", utils.FormatINR(margins.UsedMargin))
		fmt.Printf("  Total:      %s\n", utils.FormatINR(margins.AvailableMargin))
		fmt.Println()

		// Print positions
		fmt.Printf("═══ Positions (%d) ═══\n", len(positions))
		for _, p := range positions {
			pnl := p.PnL
			pnlStr := utils.FormatINR(pnl)
			if pnl >= 0 {
				pnlStr = "+" + pnlStr
			}
			fmt.Printf("  %-15s %5d @ %s  PnL: %s\n",
				p.Ticker, p.Quantity, utils.FormatINR(p.AvgPrice), pnlStr)
		}
		if len(positions) == 0 {
			fmt.Println("  No open positions")
		}
		fmt.Println()

		// Print holdings
		fmt.Printf("═══ Holdings (%d) ═══\n", len(holdings))
		for _, h := range holdings {
			fmt.Printf("  %-15s %5d @ %s  CMP: %s  PnL: %s\n",
				h.Ticker, h.Quantity, utils.FormatINR(h.AvgPrice),
				utils.FormatINR(h.LTP), utils.FormatINR(h.PnL))
		}
		if len(holdings) == 0 {
			fmt.Println("  No holdings")
		}
		fmt.Println()

		// Print recent orders
		fmt.Printf("═══ Recent Orders (%d) ═══\n", len(orders))
		for _, o := range orders {
			fmt.Printf("  [%s] %-15s %s %d @ %s  Status: %s\n",
				o.OrderID, o.Ticker, o.Side, o.Quantity, utils.FormatINR(o.Price), o.Status)
		}
		if len(orders) == 0 {
			fmt.Println("  No orders")
		}

		return nil
	},
}

func init() {
	portfolioCmd.Flags().Bool("json", false, "output result as JSON")
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
		replFlag, _ := cmd.Flags().GetBool("repl")
		nl, _ := cmd.Flags().GetString("nl")
		outputJSON, _ := cmd.Flags().GetBool("json")

		agg := datasource.NewAggregator()

		if replFlag {
			fmt.Println("📟 FinanceQL Interactive REPL")
			fmt.Println("   Type .help for commands, .quit to exit")
			fmt.Println()
			repl := financeql.NewREPL(agg)
			repl.Run()
			return nil
		}

		if nl != "" {
			fmt.Printf("🗣️  Natural Language → FinanceQL: %q\n", nl)
			fmt.Println()

			// Use LLM to translate NL to FinanceQL
			orch, err := newOrchestrator()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			prompt := fmt.Sprintf("Translate this natural language query to a FinanceQL expression. "+
				"Only return the FinanceQL expression, nothing else: %s", nl)
			result, err := orch.QuickQuery(ctx, prompt)
			if err != nil {
				return fmt.Errorf("NL translation failed: %w", err)
			}

			fqlExpr := strings.TrimSpace(result.Content)
			fmt.Printf("   Translated: %s\n", fqlExpr)
			fmt.Println()

			// Execute the translated expression
			ec := financeql.NewEvalContext(ctx, agg)
			financeql.RegisterBuiltins(ec)
			val, err := financeql.EvalQuery(ec, fqlExpr)
			if err != nil {
				return fmt.Errorf("FinanceQL execution failed: %w", err)
			}

			printFinanceQLResult(val, outputJSON)
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("provide a FinanceQL expression or use --repl")
		}

		expr := strings.Join(args, " ")
		fmt.Printf("📟 FinanceQL: %s\n", expr)
		fmt.Println()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		ec := financeql.NewEvalContext(ctx, agg)
		financeql.RegisterBuiltins(ec)
		val, err := financeql.EvalQuery(ec, expr)
		if err != nil {
			return fmt.Errorf("FinanceQL error: %w", err)
		}

		printFinanceQLResult(val, outputJSON)
		return nil
	},
}

func init() {
	queryCmd.Flags().Bool("repl", false, "start interactive FinanceQL REPL")
	queryCmd.Flags().String("nl", "", "natural language query to translate to FinanceQL")
	queryCmd.Flags().Bool("json", false, "output result as JSON")
}

// --- Chat Command ---

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat mode",
	Long:  "Start a conversational interface with the AI agent for free-form analysis queries.",
	RunE: func(cmd *cobra.Command, args []string) error {
		deep, _ := cmd.Flags().GetBool("deep")

		fmt.Println("💬 OpeNSE.ai Chat Mode")
		if deep {
			fmt.Println("   Mode: Deep Analysis (multi-agent)")
		} else {
			fmt.Println("   Mode: Quick (single-agent)")
		}
		fmt.Println("   Type 'quit' or 'exit' to leave")
		fmt.Println()

		orch, err := newOrchestrator()
		if err != nil {
			return err
		}

		if deep {
			orch.SetMode(agent.ModeMulti)
		}

		return runChatREPL(orch)
	},
}

func init() {
	chatCmd.Flags().Bool("deep", false, "use multi-agent deep analysis mode")
}

// --- Serve Command (API Server) ---

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server with embedded web UI",
	Long: `Start the HTTP REST API server for programmatic access.

The server exposes endpoints for analysis, quotes, backtesting, portfolio,
chat, FinanceQL queries, and WebSocket streaming.

By default, the embedded web UI is served at / and the API at /api/v1.
Use --no-ui to disable the web UI and serve only the API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		if port == 0 {
			port = cfg.API.Port
		}
		host, _ := cmd.Flags().GetString("host")
		if host == "" {
			host = cfg.API.Host
		}
		noUI, _ := cmd.Flags().GetBool("no-ui")

		srv, err := api.NewServer(cfg)
		if err != nil {
			return fmt.Errorf("failed to create API server: %w", err)
		}

		if noUI {
			srv.SetServeUI(false)
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		fmt.Printf("🌐 Starting OpeNSE.ai server on %s\n", addr)
		if !noUI {
			fmt.Printf("   Web UI:  http://%s/\n", resolveDisplayAddr(host, port))
		}
		fmt.Printf("   API:     http://%s/api/v1\n", resolveDisplayAddr(host, port))
		fmt.Println()
		fmt.Println("   Endpoints:")
		fmt.Println("     POST /api/v1/analyze    — run analysis")
		fmt.Println("     GET  /api/v1/quote/:t   — live quote")
		fmt.Println("     POST /api/v1/backtest   — run backtest")
		fmt.Println("     GET  /api/v1/portfolio   — portfolio summary")
		fmt.Println("     POST /api/v1/chat        — chat")
		fmt.Println("     POST /api/v1/query       — FinanceQL query")
		fmt.Println("     POST /api/v1/query/explain — explain FinanceQL")
		fmt.Println("     POST /api/v1/query/nl    — natural language query")
		fmt.Println("     GET  /api/v1/alerts      — active alerts")
		fmt.Println("     WS   /api/v1/ws          — WebSocket streaming")
		fmt.Println()
		fmt.Println("   Press Ctrl+C to stop")

		return srv.ListenAndServe(addr)
	},
}

// resolveDisplayAddr returns a display-friendly address (replaces 0.0.0.0 with localhost).
func resolveDisplayAddr(host string, port int) string {
	if host == "" || host == "0.0.0.0" {
		return fmt.Sprintf("localhost:%d", port)
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func init() {
	serveCmd.Flags().IntP("port", "p", 0, "server port (default from config)")
	serveCmd.Flags().String("host", "", "server host (default from config)")
	serveCmd.Flags().Bool("no-ui", false, "disable embedded web UI (API only)")
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

// ============================================================
// Helper functions
// ============================================================

func printAgentResult(r *agent.AgentResult) {
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  Agent: %s (%s)\n", r.AgentName, r.Role)
	fmt.Printf("  Duration: %s\n", r.Duration.Round(time.Millisecond))
	fmt.Println("═══════════════════════════════════════")
	fmt.Println()
	fmt.Println(r.Content)
	fmt.Println()

	if r.Analysis != nil {
		fmt.Printf("  Recommendation: %s\n", r.Analysis.Recommendation)
		fmt.Printf("  Confidence:     %.0f%%\n", float64(r.Analysis.Confidence)*100)
		if len(r.Analysis.Signals) > 0 {
			fmt.Println("  Signals:")
			for _, sig := range r.Analysis.Signals {
				fmt.Printf("    [%s] %s — %s (%.0f%%)\n",
					sig.Source, sig.Type, sig.Reason, float64(sig.Confidence)*100)
			}
		}
	}

	if r.ToolCalls > 0 {
		fmt.Printf("\n  Tool Calls: %d\n", r.ToolCalls)
	}
	if r.Tokens > 0 {
		fmt.Printf("  Tokens:     %d\n", r.Tokens)
	}
}

func printBacktestResult(r *models.BacktestResult) {
	fmt.Println("═══════════════════════════════════════")
	fmt.Println("  Backtest Results")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("  Strategy:       %s\n", r.StrategyName)
	fmt.Printf("  Ticker:         %s\n", r.Ticker)
	fmt.Printf("  Period:         %s to %s\n",
		r.From.Format("2006-01-02"), r.To.Format("2006-01-02"))
	fmt.Printf("  Initial:        %s\n", utils.FormatINR(r.InitialCapital))
	fmt.Printf("  Final:          %s\n", utils.FormatINR(r.FinalCapital))
	fmt.Println()
	fmt.Printf("  Total Return:   %s\n", utils.FormatPct(r.TotalReturnPct))
	fmt.Printf("  CAGR:           %s\n", utils.FormatPct(r.CAGR))
	fmt.Printf("  Sharpe Ratio:   %.2f\n", r.SharpeRatio)
	fmt.Printf("  Sortino Ratio:  %.2f\n", r.SortinoRatio)
	fmt.Printf("  Max Drawdown:   %s\n", utils.FormatPct(r.MaxDrawdownPct))
	fmt.Println()
	fmt.Printf("  Total Trades:   %d\n", r.TotalTrades)
	fmt.Printf("  Win Rate:       %s\n", utils.FormatPct(r.WinRate))
	fmt.Printf("  Profit Factor:  %.2f\n", r.ProfitFactor)
	fmt.Println("═══════════════════════════════════════")
}

func printFinanceQLResult(val financeql.Value, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		switch val.Type {
		case financeql.TypeScalar:
			_ = enc.Encode(map[string]any{"type": "scalar", "value": val.Scalar})
		case financeql.TypeString:
			_ = enc.Encode(map[string]any{"type": "string", "value": val.Str})
		case financeql.TypeBool:
			_ = enc.Encode(map[string]any{"type": "bool", "value": val.Bool})
		case financeql.TypeVector:
			_ = enc.Encode(map[string]any{"type": "vector", "points": val.Vector})
		case financeql.TypeMatrix:
			_ = enc.Encode(map[string]any{"type": "matrix", "series": val.Matrix})
		case financeql.TypeTable:
			_ = enc.Encode(map[string]any{"type": "table", "rows": val.Table})
		default:
			_ = enc.Encode(map[string]any{"type": "nil"})
		}
		return
	}

	switch val.Type {
	case financeql.TypeScalar:
		fmt.Printf("  Result: %.4f\n", val.Scalar)
	case financeql.TypeString:
		fmt.Printf("  Result: %s\n", val.Str)
	case financeql.TypeBool:
		fmt.Printf("  Result: %v\n", val.Bool)
	case financeql.TypeVector:
		fmt.Printf("  Vector (%d points):\n", len(val.Vector))
		start := 0
		if len(val.Vector) > 10 {
			start = len(val.Vector) - 10
			fmt.Printf("  ... showing last 10 of %d\n", len(val.Vector))
		}
		for _, pt := range val.Vector[start:] {
			fmt.Printf("    %s  %.4f\n", pt.Time.Format("2006-01-02"), pt.Value)
		}
	case financeql.TypeMatrix:
		for name, pts := range val.Matrix {
			fmt.Printf("  Series: %s (%d points)\n", name, len(pts))
			start := 0
			if len(pts) > 5 {
				start = len(pts) - 5
			}
			for _, pt := range pts[start:] {
				fmt.Printf("    %s  %.4f\n", pt.Time.Format("2006-01-02"), pt.Value)
			}
		}
	case financeql.TypeTable:
		if len(val.Table) == 0 {
			fmt.Println("  Empty table")
			return
		}
		var keys []string
		for k := range val.Table[0] {
			keys = append(keys, k)
		}
		fmt.Printf("  Table (%d rows):\n", len(val.Table))
		header := "    "
		for _, k := range keys {
			header += fmt.Sprintf("%-15s", k)
		}
		fmt.Println(header)
		fmt.Println("    " + strings.Repeat("─", len(keys)*15))
		limit := len(val.Table)
		if limit > 20 {
			limit = 20
		}
		for _, row := range val.Table[:limit] {
			line := "    "
			for _, k := range keys {
				line += fmt.Sprintf("%-15v", row[k])
			}
			fmt.Println(line)
		}
		if len(val.Table) > 20 {
			fmt.Printf("    ... and %d more rows\n", len(val.Table)-20)
		}
	default:
		fmt.Println("  Result: nil")
	}
}

func buildCompositeAnalysis(ticker string, result *agent.AgentResult) *models.CompositeAnalysis {
	ca := &models.CompositeAnalysis{
		Ticker:    ticker,
		Summary:   result.Content,
		Timestamp: time.Now(),
		Timeframe: "medium-term",
	}
	if result.Analysis != nil {
		ca.Recommendation = result.Analysis.Recommendation
		ca.Confidence = result.Analysis.Confidence
	}
	return ca
}

func findStrategy(name string) backtest.Strategy {
	name = strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, s := range backtest.BuiltinStrategies() {
		sName := strings.ToLower(strings.ReplaceAll(s.Name(), " ", "_"))
		if sName == name || strings.Contains(sName, name) {
			return s
		}
	}
	return nil
}

func listStrategyNames() []string {
	var names []string
	for _, s := range backtest.BuiltinStrategies() {
		names = append(names, s.Name())
	}
	return names
}

func printWatchlist(ctx context.Context, agg *datasource.Aggregator, tickers []string) {
	fmt.Printf("\033[2J\033[H") // clear screen
	fmt.Printf("  %-15s %12s %10s %10s   %s\n", "TICKER", "PRICE", "CHANGE", "CHANGE%", "TIME")
	fmt.Println("  " + strings.Repeat("─", 65))

	for _, t := range tickers {
		quote, err := agg.YFinance().GetQuote(ctx, t)
		if err != nil {
			fmt.Printf("  %-15s  ⚠ error: %s\n", t, err)
			continue
		}
		changeStr := utils.FormatINR(quote.Change)
		if quote.Change >= 0 {
			changeStr = "+" + changeStr
		}
		fmt.Printf("  %-15s %12s %10s %10s   %s\n",
			t,
			utils.FormatINR(quote.LastPrice),
			changeStr,
			utils.FormatPct(quote.ChangePct),
			quote.Timestamp.Format("15:04:05"),
		)
	}
	fmt.Printf("\n  Last updated: %s\n", utils.FormatDateTimeIST(utils.NowIST()))
}

func runChatREPL(orch *agent.Orchestrator) error {
	var history []llm.Message
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("you> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println("👋 Goodbye!")
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := orch.Chat(ctx, input, history)
		cancel()
		if err != nil {
			fmt.Printf("❌ Error: %s\n\n", err)
			continue
		}

		fmt.Printf("\n🤖 %s:\n%s\n\n", result.AgentName, result.Content)

		// Append to history
		history = append(history, llm.UserMessage(input))
		history = append(history, llm.AssistantMessage(result.Content))

		// Keep history manageable
		if len(history) > 20 {
			history = history[len(history)-20:]
		}
	}
	return nil
}

func runTradeREPL(ctx context.Context, rm *broker.RiskManager) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("trade> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "quit", "exit":
			fmt.Println("👋 Goodbye!")
			return nil

		case "positions":
			positions, err := rm.GetPositions(ctx)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}
			fmt.Printf("Open positions: %d\n", len(positions))
			for _, p := range positions {
				fmt.Printf("  %-15s %5d @ %s  PnL: %s\n",
					p.Ticker, p.Quantity, utils.FormatINR(p.AvgPrice), utils.FormatINR(p.PnL))
			}

		case "orders":
			orders, err := rm.GetOrders(ctx)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}
			fmt.Printf("Orders: %d\n", len(orders))
			for _, o := range orders {
				fmt.Printf("  [%s] %-15s %s %d @ %s  %s\n",
					o.OrderID, o.Ticker, o.Side, o.Quantity, utils.FormatINR(o.Price), o.Status)
			}

		case "margins":
			m, err := rm.GetMargins(ctx)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				continue
			}
			fmt.Printf("  Available: %s\n  Used: %s\n  Total: %s\n",
				utils.FormatINR(m.AvailableCash), utils.FormatINR(m.UsedMargin), utils.FormatINR(m.AvailableMargin))

		case "buy", "sell":
			if len(parts) < 4 {
				fmt.Printf("Usage: %s TICKER QUANTITY PRICE\n", cmd)
				continue
			}
			ticker := utils.NormalizeTicker(parts[1])
			var qty int
			var price float64
			fmt.Sscanf(parts[2], "%d", &qty)
			fmt.Sscanf(parts[3], "%f", &price)

			side := models.Buy
			if cmd == "sell" {
				side = models.Sell
			}

			req := models.OrderRequest{
				Ticker:    ticker,
				Side:      side,
				Quantity:  qty,
				Price:     price,
				OrderType: models.Limit,
				Product:   models.CNC,
			}

			resp, err := rm.PlaceOrder(ctx, req)
			if err != nil {
				fmt.Printf("❌ Order failed: %v\n", err)
				continue
			}
			fmt.Printf("✅ Order placed: %s (%s)\n", resp.OrderID, resp.Status)

		case "cancel":
			if len(parts) < 2 {
				fmt.Println("Usage: cancel ORDER_ID")
				continue
			}
			if err := rm.CancelOrder(ctx, parts[1]); err != nil {
				fmt.Printf("❌ Cancel failed: %v\n", err)
				continue
			}
			fmt.Println("✅ Order cancelled")

		default:
			fmt.Println("Unknown command. Available: buy, sell, positions, orders, margins, cancel, quit")
		}
		fmt.Println()
	}
	return nil
}
