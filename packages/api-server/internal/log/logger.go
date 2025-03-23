package log

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
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
	logger := &Logger{
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

	return logger
}

func (l *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Info(msg)
}

func (l *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Error(msg)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Fatal(msg)
}

func (l *Logger) Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.Logger.Debug(msg)
}
