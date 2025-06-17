package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging with multiple outputs
type Logger struct {
	level       LogLevel
	mainLogger  *log.Logger
	auditLogger *log.Logger
	config      *config.Config
	mu          sync.RWMutex
	sessionID   string
}

var (
	globalLogger *Logger
	loggerOnce   sync.Once
)

// InitLogger initializes the global logger
func InitLogger(cfg *config.Config, sessionID string) error {
	var err error
	loggerOnce.Do(func() {
		globalLogger, err = NewLogger(cfg, sessionID)
	})
	return err
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Fallback to basic logger if not initialized
		return &Logger{
			level:      INFO,
			mainLogger: log.New(os.Stdout, "[COHORT] ", log.LstdFlags|log.Lshortfile),
			sessionID:  "default",
		}
	}
	return globalLogger
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config, sessionID string) (*Logger, error) {
	logger := &Logger{
		level:     parseLogLevel(cfg.Logging.Level),
		config:    cfg,
		sessionID: sessionID,
	}

	// Setup main logger
	var mainWriter io.Writer = os.Stdout
	if cfg.Logging.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.File), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(cfg.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		mainWriter = file
	}

	logger.mainLogger = log.New(mainWriter, fmt.Sprintf("[COHORT-%s] ", sessionID),
		log.LstdFlags|log.Lshortfile)

	// Setup audit logger if enabled
	if cfg.Logging.EnableAudit {
		auditFile := cfg.Logging.AuditFile
		if auditFile == "" {
			auditFile = "audit.log"
		}

		if err := os.MkdirAll(filepath.Dir(auditFile), 0755); err != nil {
			return nil, fmt.Errorf("failed to create audit log directory: %w", err)
		}

		file, err := os.OpenFile(auditFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open audit log file: %w", err)
		}

		logger.auditLogger = log.New(file, fmt.Sprintf("[AUDIT-%s] ", sessionID),
			log.LstdFlags|log.Lshortfile)
	}

	return logger, nil
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.log(DEBUG, format, args...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= INFO {
		l.log(INFO, format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= WARN {
		l.log(WARN, format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= ERROR {
		l.log(ERROR, format, args...)
	}
}

// Audit logs a security audit event
func (l *Logger) Audit(event string, details map[string]interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	message := fmt.Sprintf("AUDIT_EVENT=%s TIMESTAMP=%s SESSION=%s",
		event, timestamp, l.sessionID)

	for key, value := range details {
		message += fmt.Sprintf(" %s=%v", key, value)
	}

	if l.auditLogger != nil {
		l.auditLogger.Println(message)
	}

	// Also log audit events to main logger at WARN level
	if l.level <= WARN {
		l.mainLogger.Printf("[AUDIT] %s", message)
	}
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	levelStr := levelToString(level)
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s", levelStr, message)

	if l.mainLogger != nil {
		l.mainLogger.Print(logLine)
	}
}

// levelToString converts LogLevel to string
func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Close closes all log outputs
func (l *Logger) Close() error {
	return nil
}

// Helper functions for global logger
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Audit(event string, details map[string]interface{}) {
	GetLogger().Audit(event, details)
}
