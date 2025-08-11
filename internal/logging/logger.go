package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// Logger represents the enhanced logging interface
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	
	WithContext(ctx context.Context) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithComponent(component string) Logger
	
	LogAPICall(method, path string, statusCode int, duration time.Duration, fields ...Field)
	LogUserAction(userID, action string, fields ...Field)
	LogPerformance(operation string, duration time.Duration, fields ...Field)
	LogSecurity(event string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Fields represents multiple log fields
type Fields map[string]interface{}

// Config represents the logging configuration
type Config struct {
	Level      LogLevel `json:"level" yaml:"level"`
	Format     string   `json:"format" yaml:"format"` // json, text
	Output     string   `json:"output" yaml:"output"` // stdout, stderr, file
	FilePath   string   `json:"file_path" yaml:"file_path"`
	MaxSize    int      `json:"max_size" yaml:"max_size"`       // MB
	MaxBackups int      `json:"max_backups" yaml:"max_backups"` // number of backup files
	MaxAge     int      `json:"max_age" yaml:"max_age"`         // days
	Compress   bool     `json:"compress" yaml:"compress"`
	
	// Component-specific settings
	EnableAPILogging         bool `json:"enable_api_logging" yaml:"enable_api_logging"`
	EnableUserActionLogging  bool `json:"enable_user_action_logging" yaml:"enable_user_action_logging"`
	EnablePerformanceLogging bool `json:"enable_performance_logging" yaml:"enable_performance_logging"`
	EnableSecurityLogging    bool `json:"enable_security_logging" yaml:"enable_security_logging"`
	EnableStackTrace         bool `json:"enable_stack_trace" yaml:"enable_stack_trace"`
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() *Config {
	return &Config{
		Level:                    InfoLevel,
		Format:                   "json",
		Output:                   "stdout",
		MaxSize:                  100,
		MaxBackups:               3,
		MaxAge:                   28,
		Compress:                 true,
		EnableAPILogging:         true,
		EnableUserActionLogging:  true,
		EnablePerformanceLogging: true,
		EnableSecurityLogging:    true,
		EnableStackTrace:         false,
	}
}

// enhancedLogger implements the Logger interface
type enhancedLogger struct {
	logger    *logrus.Logger
	config    *Config
	component string
	fields    Fields
	context   context.Context
}

// NewLogger creates a new enhanced logger
func NewLogger(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	logger := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(string(config.Level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", config.Level)
	}
	logger.SetLevel(level)

	// Set formatter
	switch config.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		return nil, fmt.Errorf("unsupported log format: %s", config.Format)
	}

	// Set output
	switch config.Output {
	case "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "file":
		if config.FilePath == "" {
			return nil, fmt.Errorf("file path required for file output")
		}
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger.SetOutput(file)
	default:
		return nil, fmt.Errorf("unsupported log output: %s", config.Output)
	}

	// Enable caller reporting for stack traces
	if config.EnableStackTrace {
		logger.SetReportCaller(true)
	}

	return &enhancedLogger{
		logger: logger,
		config: config,
		fields: make(Fields),
	}, nil
}

// Debug logs a debug message
func (l *enhancedLogger) Debug(msg string, fields ...Field) {
	l.log(logrus.DebugLevel, msg, fields...)
}

// Info logs an info message
func (l *enhancedLogger) Info(msg string, fields ...Field) {
	l.log(logrus.InfoLevel, msg, fields...)
}

// Warn logs a warning message
func (l *enhancedLogger) Warn(msg string, fields ...Field) {
	l.log(logrus.WarnLevel, msg, fields...)
}

// Error logs an error message
func (l *enhancedLogger) Error(msg string, fields ...Field) {
	l.log(logrus.ErrorLevel, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *enhancedLogger) Fatal(msg string, fields ...Field) {
	l.log(logrus.FatalLevel, msg, fields...)
}

// log is the internal logging method
func (l *enhancedLogger) log(level logrus.Level, msg string, fields ...Field) {
	entry := l.logger.WithFields(l.buildFields(fields...))
	
	if l.component != "" {
		entry = entry.WithField("component", l.component)
	}
	
	if l.context != nil {
		// Add context values if available
		if requestID := l.context.Value("request_id"); requestID != nil {
			entry = entry.WithField("request_id", requestID)
		}
		if userID := l.context.Value("user_id"); userID != nil {
			entry = entry.WithField("user_id", userID)
		}
		if sessionID := l.context.Value("session_id"); sessionID != nil {
			entry = entry.WithField("session_id", sessionID)
		}
	}

	// Add caller information for errors and above
	if level >= logrus.ErrorLevel && l.config.EnableStackTrace {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry = entry.WithFields(logrus.Fields{
				"caller_file":     file,
				"caller_line":     line,
				"caller_function": runtime.FuncForPC(pc).Name(),
			})
		}
	}

	entry.Log(level, msg)
}

// buildFields converts Field slice to logrus.Fields
func (l *enhancedLogger) buildFields(fields ...Field) logrus.Fields {
	result := make(logrus.Fields)
	
	// Add existing fields
	for k, v := range l.fields {
		result[k] = v
	}
	
	// Add new fields
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	
	return result
}

// WithContext returns a logger with context
func (l *enhancedLogger) WithContext(ctx context.Context) Logger {
	return &enhancedLogger{
		logger:    l.logger,
		config:    l.config,
		component: l.component,
		fields:    l.fields,
		context:   ctx,
	}
}

// WithField returns a logger with an additional field
func (l *enhancedLogger) WithField(key string, value interface{}) Logger {
	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value
	
	return &enhancedLogger{
		logger:    l.logger,
		config:    l.config,
		component: l.component,
		fields:    newFields,
		context:   l.context,
	}
}

// WithFields returns a logger with additional fields
func (l *enhancedLogger) WithFields(fields Fields) Logger {
	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	
	return &enhancedLogger{
		logger:    l.logger,
		config:    l.config,
		component: l.component,
		fields:    newFields,
		context:   l.context,
	}
}

// WithComponent returns a logger with a component name
func (l *enhancedLogger) WithComponent(component string) Logger {
	return &enhancedLogger{
		logger:    l.logger,
		config:    l.config,
		component: component,
		fields:    l.fields,
		context:   l.context,
	}
}

// LogAPICall logs an API call with standardized fields
func (l *enhancedLogger) LogAPICall(method, path string, statusCode int, duration time.Duration, fields ...Field) {
	if !l.config.EnableAPILogging {
		return
	}

	allFields := append(fields,
		Field{"method", method},
		Field{"path", path},
		Field{"status_code", statusCode},
		Field{"duration_ms", duration.Milliseconds()},
		Field{"type", "api_call"},
	)

	level := logrus.InfoLevel
	if statusCode >= 400 {
		level = logrus.ErrorLevel
	}

	l.log(level, fmt.Sprintf("API %s %s", method, path), allFields...)
}

// LogUserAction logs a user action
func (l *enhancedLogger) LogUserAction(userID, action string, fields ...Field) {
	if !l.config.EnableUserActionLogging {
		return
	}

	allFields := append(fields,
		Field{"user_id", userID},
		Field{"action", action},
		Field{"type", "user_action"},
	)

	l.log(logrus.InfoLevel, fmt.Sprintf("User action: %s", action), allFields...)
}

// LogPerformance logs performance metrics
func (l *enhancedLogger) LogPerformance(operation string, duration time.Duration, fields ...Field) {
	if !l.config.EnablePerformanceLogging {
		return
	}

	allFields := append(fields,
		Field{"operation", operation},
		Field{"duration_ms", duration.Milliseconds()},
		Field{"type", "performance"},
	)

	level := logrus.InfoLevel
	if duration > time.Second {
		level = logrus.WarnLevel
	}

	l.log(level, fmt.Sprintf("Performance: %s", operation), allFields...)
}

// LogSecurity logs security-related events
func (l *enhancedLogger) LogSecurity(event string, fields ...Field) {
	if !l.config.EnableSecurityLogging {
		return
	}

	allFields := append(fields,
		Field{"event", event},
		Field{"type", "security"},
	)

	l.log(logrus.WarnLevel, fmt.Sprintf("Security event: %s", event), allFields...)
}

// Helper functions for creating fields
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

func Duration(key string, duration time.Duration) Field {
	return Field{Key: key, Value: duration.Milliseconds()}
}

// JSON helper for complex objects
func JSON(key string, obj interface{}) Field {
	data, _ := json.Marshal(obj)
	return Field{Key: key, Value: string(data)}
}

// Global logger instance
var globalLogger Logger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(config *Config) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		// Initialize with default config if not set
		globalLogger, _ = NewLogger(DefaultConfig())
	}
	return globalLogger
}

// Convenience functions using global logger
func Debug(msg string, fields ...Field) {
	GetGlobalLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetGlobalLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetGlobalLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	GetGlobalLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetGlobalLogger().Fatal(msg, fields...)
}
