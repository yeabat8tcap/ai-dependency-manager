package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init initializes the logger with the specified level and format
func Init(level, format string) {
	log = logrus.New()
	
	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	
	// Set log format
	switch format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
	
	log.SetOutput(os.Stdout)
}

// Debug logs a debug message
func Debug(format string, args ...interface{}) {
	if log != nil {
		log.Debugf(format, args...)
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if log != nil {
		log.Infof(format, args...)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if log != nil {
		log.Warnf(format, args...)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if log != nil {
		log.Errorf(format, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(format string, args ...interface{}) {
	if log != nil {
		log.Fatalf(format, args...)
	}
}

// WithField adds a field to the logger
func WithField(key string, value interface{}) *logrus.Entry {
	if log != nil {
		return log.WithField(key, value)
	}
	return logrus.NewEntry(logrus.New())
}

// WithFields adds multiple fields to the logger
func WithFields(fields logrus.Fields) *logrus.Entry {
	if log != nil {
		return log.WithFields(fields)
	}
	return logrus.NewEntry(logrus.New())
}
