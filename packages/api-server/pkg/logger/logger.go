package logger

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

var (
	// logger is the global logger instance
	logger *Logger
	once   sync.Once
)

// Logger wraps the standard logger with color support
type Logger struct {
	*logrus.Logger
	green  *color.Color
	cyan   *color.Color
	red    *color.Color
	blue   *color.Color
	yellow *color.Color
	bold   *color.Color
}

// New creates a new logger with color support
func New() *Logger {
	once.Do(func() {
		logger = &Logger{
			Logger: logrus.New(),
			green:  color.New(color.FgGreen),
			cyan:   color.New(color.FgCyan),
			red:    color.New(color.FgRed),
			blue:   color.New(color.FgBlue),
			yellow: color.New(color.FgYellow),
			bold:   color.New(color.Bold),
		}

		// Configure logrus
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006/01/02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     true,
			DisableSorting:  true,
		})

		// Set log level based on environment variable
		if os.Getenv("DEBUG") == "true" {
			logger.SetLevel(logrus.DebugLevel)
			logger.Info("Debug logging enabled")
		} else {
			logger.SetLevel(logrus.InfoLevel)
		}
	})
	return logger
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Debug(msg)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Info(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Warn(msg)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Error(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Fatal(msg)
}

// IsDebugEnabled returns whether debug logging is enabled
func (l *Logger) IsDebugEnabled() bool {
	return l.GetLevel() == logrus.DebugLevel
}
