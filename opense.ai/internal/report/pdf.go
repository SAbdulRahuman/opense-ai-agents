package report

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ════════════════════════════════════════════════════════════════════
// PDF Generator — HTML → PDF via wkhtmltopdf / chromium headless
// ════════════════════════════════════════════════════════════════════

// PDFEngine specifies which engine to use for HTML→PDF conversion.
type PDFEngine string

const (
	EngineWKHTML    PDFEngine = "wkhtmltopdf"
	EngineChromium  PDFEngine = "chromium"
	EngineNone      PDFEngine = "none" // skip PDF, return HTML
)

// PDFConfig holds configuration for PDF generation.
type PDFConfig struct {
	Engine     PDFEngine // default: auto-detect
	PageSize   string    // default: "A4"
	Orientation string   // "portrait" (default) or "landscape"
	MarginTop  string    // default: "15mm"
	MarginBottom string  // default: "15mm"
	MarginLeft string    // default: "10mm"
	MarginRight string   // default: "10mm"
	OutputPath string    // required: output PDF file path
}

// DefaultPDFConfig returns sensible defaults for PDF generation.
func DefaultPDFConfig() PDFConfig {
	return PDFConfig{
		Engine:      EngineNone, // auto-detect
		PageSize:    "A4",
		Orientation: "portrait",
		MarginTop:   "15mm",
		MarginBottom: "15mm",
		MarginLeft:  "10mm",
		MarginRight: "10mm",
	}
}

// DetectPDFEngine checks which PDF engine is available on the system.
func DetectPDFEngine() PDFEngine {
	if _, err := exec.LookPath("wkhtmltopdf"); err == nil {
		return EngineWKHTML
	}
	for _, name := range []string{"chromium-browser", "chromium", "google-chrome", "google-chrome-stable"} {
		if _, err := exec.LookPath(name); err == nil {
			return EngineChromium
		}
	}
	return EngineNone
}

// GeneratePDF converts an HTML string to a PDF file.
// It writes the HTML to a temp file, runs the conversion engine, and
// writes the output PDF to cfg.OutputPath.
func GeneratePDF(html string, cfg PDFConfig) error {
	if cfg.OutputPath == "" {
		return fmt.Errorf("output path is required")
	}

	engine := cfg.Engine
	if engine == "" || engine == EngineNone {
		engine = DetectPDFEngine()
	}

	switch engine {
	case EngineWKHTML:
		return generateWithWKHTML(html, cfg)
	case EngineChromium:
		return generateWithChromium(html, cfg)
	case EngineNone:
		// No engine available — write HTML as fallback
		return writeHTMLFallback(html, cfg.OutputPath)
	default:
		return fmt.Errorf("unsupported PDF engine: %s", engine)
	}
}

func generateWithWKHTML(html string, cfg PDFConfig) error {
	tmpFile, err := writeTempHTML(html)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	args := []string{
		"--page-size", cfg.PageSize,
		"--orientation", cfg.Orientation,
		"--margin-top", cfg.MarginTop,
		"--margin-bottom", cfg.MarginBottom,
		"--margin-left", cfg.MarginLeft,
		"--margin-right", cfg.MarginRight,
		"--encoding", "UTF-8",
		"--enable-local-file-access",
		"--quiet",
		tmpFile,
		cfg.OutputPath,
	}

	cmd := exec.Command("wkhtmltopdf", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("wkhtmltopdf failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func generateWithChromium(html string, cfg PDFConfig) error {
	tmpFile, err := writeTempHTML(html)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// Find available chromium binary
	var chromiumBin string
	for _, name := range []string{"chromium-browser", "chromium", "google-chrome", "google-chrome-stable"} {
		if path, err := exec.LookPath(name); err == nil {
			chromiumBin = path
			break
		}
	}
	if chromiumBin == "" {
		return fmt.Errorf("chromium not found in PATH")
	}

	// Ensure output is absolute path
	absOutput, err := filepath.Abs(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("resolving output path: %w", err)
	}

	args := []string{
		"--headless",
		"--disable-gpu",
		"--no-sandbox",
		"--print-to-pdf=" + absOutput,
		"--print-to-pdf-no-header",
	}

	// Page size
	if strings.EqualFold(cfg.Orientation, "landscape") {
		args = append(args, "--landscape")
	}

	args = append(args, "file://"+tmpFile)

	cmd := exec.Command(chromiumBin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("chromium PDF export failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func writeTempHTML(html string) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "openseai_report.html")
	if err := os.WriteFile(tmpFile, []byte(html), 0644); err != nil {
		return "", fmt.Errorf("writing temp HTML: %w", err)
	}
	return tmpFile, nil
}

func writeHTMLFallback(html string, outputPath string) error {
	// Change extension to .html if .pdf was specified
	if strings.HasSuffix(strings.ToLower(outputPath), ".pdf") {
		outputPath = outputPath[:len(outputPath)-4] + ".html"
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("writing HTML fallback: %w", err)
	}
	return nil
}

// IsPDFSupported returns true if a PDF engine is available.
func IsPDFSupported() bool {
	return DetectPDFEngine() != EngineNone
}
