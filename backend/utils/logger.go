package utils

import (
	"fmt"
	"log"
	"os"
)

// Logger provides structured logging with component prefixes
type Logger struct {
	component string
}

// NewLogger creates a new logger with the specified component name
func NewLogger(component string) *Logger {
	return &Logger{component: component}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	log.Printf("[INFO] %s %s", l.component, fmt.Sprintf(format, v...))
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	log.Printf("[ERROR] %s %s", l.component, fmt.Sprintf(format, v...))
}

// Debug logs a debug message (only shown when DEBUG=1)
func (l *Logger) Debug(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "1" || os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] %s %s", l.component, fmt.Sprintf(format, v...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	log.Printf("[WARN] %s %s", l.component, fmt.Sprintf(format, v...))
}

// Predefined loggers for different components
var (
	LogDB       = NewLogger("[DB]")
	LogHandler  = NewLogger("[HANDLER]")
	LogRepo     = NewLogger("[REPO]")
	LogAuth     = NewLogger("[AUTH]")
	LogTemplate = NewLogger("[TEMPLATE]")
)
