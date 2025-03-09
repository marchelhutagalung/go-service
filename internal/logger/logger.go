package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Init initializes the logger with the specified environment configuration
func Init(env string) {
	// Set default formatter
	if env == "production" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			ForceColors:     true,
		})
	}

	// Set output
	log.SetOutput(os.Stdout)

	// Set log level based on environment
	if env == "development" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

// Field creates a logrus field for structured logging
func Field(key string, value interface{}) logrus.Fields {
	return logrus.Fields{
		key: value,
	}
}

// WithFields creates multiple fields for structured logging
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return log.WithFields(logrus.Fields(fields))
}

// Debug logs a debug message
func Debug(message string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		mergedFields := mergeFields(fields...)
		log.WithFields(mergedFields).Debug(message)
	} else {
		log.Debug(message)
	}
}

// Info logs an informational message
func Info(message string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		mergedFields := mergeFields(fields...)
		log.WithFields(mergedFields).Info(message)
	} else {
		log.Info(message)
	}
}

// Warn logs a warning message
func Warn(message string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		mergedFields := mergeFields(fields...)
		log.WithFields(mergedFields).Warn(message)
	} else {
		log.Warn(message)
	}
}

// Error logs an error message
func Error(message string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		mergedFields := mergeFields(fields...)
		log.WithFields(mergedFields).Error(message)
	} else {
		log.Error(message)
	}
}

// Fatal logs a fatal message and exits the application
func Fatal(message string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		mergedFields := mergeFields(fields...)
		log.WithFields(mergedFields).Fatal(message)
	} else {
		log.Fatal(message)
	}
}

// Helper function to merge multiple fields
func mergeFields(fieldsArray ...logrus.Fields) logrus.Fields {
	result := logrus.Fields{}
	for _, fields := range fieldsArray {
		for k, v := range fields {
			result[k] = v
		}
	}
	return result
}

// GetLogger returns the underlying logrus logger
func GetLogger() *logrus.Logger {
	return log
}
