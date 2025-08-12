package logging

import (
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"testing"
)

// LoggerOptions configures logger behavior
type LoggerOptions struct {
	Format      string       // "json" or "text" (default)
	Level       slog.Leveler // Min log level
	AddSource   bool         // Include source location
	Output      io.Writer    // Output destination (default: os.Stdout)
	WithPID     bool         // Include process ID
	WithVersion bool         // Include Go version
	WithBuild   bool         // Include build info
	Attributes  []slog.Attr  // Additional attributes
}

var DefaultOptions = LoggerOptions{
	Format:      "text",
	Level:       slog.LevelInfo,
	Output:      os.Stdout,
	WithPID:     true,
	WithVersion: true,
}

// New creates a configured logger
func New(opts LoggerOptions) *slog.Logger {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	}

	var handler slog.Handler
	switch opts.Format {
	case "json":
		handler = slog.NewJSONHandler(opts.Output, handlerOpts)
	default:
		handler = slog.NewTextHandler(opts.Output, handlerOpts)
	}

	logger := slog.New(handler)

	// Convert attributes to interface{} slice
	attrs := make([]any, 0, len(opts.Attributes)+3) // Pre-allocate space

	if opts.WithPID {
		attrs = append(attrs, slog.Int("pid", os.Getpid()))
	}

	if buildInfo, _ := debug.ReadBuildInfo(); buildInfo != nil {
		if opts.WithVersion {
			attrs = append(attrs, slog.String("version", buildInfo.GoVersion))
		}
		if opts.WithBuild {
			attrs = append(attrs, slog.Group("build",
				slog.String("path", buildInfo.Main.Path),
				slog.String("version", buildInfo.Main.Version),
			))
		}
	}

	// Add custom attributes
	for _, attr := range opts.Attributes {
		attrs = append(attrs, attr)
	}

	if len(attrs) > 0 {
		logger = logger.With(attrs...)
	}

	return logger
}

// SetDefault configures the global logger
func SetDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}

// NewDefault creates logger with default options
func NewDefault() *slog.Logger {
	return New(DefaultOptions)
}

// Shortcut functions for common scenarios

// NewProduction creates a production-ready JSON logger with:
// - JSON formatting
// - Info level logging
// - Source locations
// - Process ID
// - Version information
// - No build info by default (to avoid sensitive info in prod)
func NewProduction() *slog.Logger {
	return New(LoggerOptions{
		Format:      "json",
		Level:       slog.LevelInfo,
		AddSource:   true,
		WithPID:     true,
		WithVersion: true,
		WithBuild:   false, // Disabled by default in production
	})
}

// NewDevelopment creates a development-friendly logger with:
// - Human-readable text format
// - Debug level logging
// - Source locations
// - Process ID
// - Version information
// - Build info (helpful during development)
func NewDevelopment() *slog.Logger {
	return New(LoggerOptions{
		Format:      "text",
		Level:       slog.LevelDebug,
		AddSource:   true,
		WithPID:     true,
		WithVersion: true,
		WithBuild:   true, // Enabled for development debugging
	})
}

// NewDiscard creates a no-op logger that discards all log output.
// Useful for:
// - Tests where you want to suppress log output
// - Benchmarking to eliminate logging overhead
// - Disabling logging in specific components
func NewDiscard() *slog.Logger {
	return New(LoggerOptions{
		Output:    io.Discard,
		Level:     slog.LevelError + 1, // Higher than any standard level
		AddSource: false,
		WithPID:   false,
	})
}

// ===== usahe

func ExampleTestSomething(t *testing.T) {
	// Use discard logger in tests
	testLogger := NewDiscard()
	SetDefault(testLogger)

	// These won't produce any output
	slog.Info("Test starting")
	slog.Error("Simulated error")
}

func ExampleSetupDevelopment() {
	// Development logger (text format, debug level)
	devLogger := NewDevelopment()
	SetDefault(devLogger)

	// Debug logs will be visible
	slog.Debug("Loading configuration",
		"path", "./config/dev.yaml",
		"attempt", 1)
}

func ExampleSetupProduction() {
	// Production logger (JSON format, info level)
	prodLogger := NewProduction()
	SetDefault(prodLogger)

	// Example structured logging
	slog.Info("Server starting",
		"port", 8080,
		"environment", "production")
}

func ExampleBasicUsage() {
	// Simple default logger (text format, stdout, info level)
	logger := NewDefault()
	logger.Info("Application started")

	// Set as default logger for package-level logging
	SetDefault(logger)
	slog.Info("Now using default logger")
}
