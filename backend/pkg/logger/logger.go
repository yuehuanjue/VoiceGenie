package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger
)

// Init initializes the logger
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
	case "panic":
		log.SetLevel(logrus.PanicLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	if format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output to stdout
	log.SetOutput(os.Stdout)

	// Add hook for file logging if needed
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(io.MultiWriter(os.Stdout, file))
		}
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	if log == nil {
		return
	}
	log.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	log.Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	if log == nil {
		fmt.Println(args...)
		return
	}
	log.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	if log == nil {
		fmt.Printf(format+"\n", args...)
		return
	}
	log.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	if log == nil {
		fmt.Println(args...)
		return
	}
	log.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	if log == nil {
		fmt.Printf(format+"\n", args...)
		return
	}
	log.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	if log == nil {
		fmt.Println(args...)
		return
	}
	log.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	if log == nil {
		fmt.Printf(format+"\n", args...)
		return
	}
	log.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	if log == nil {
		fmt.Println(args...)
		os.Exit(1)
	}
	log.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	if log == nil {
		fmt.Printf(format+"\n", args...)
		os.Exit(1)
	}
	log.Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	if log == nil {
		panic(fmt.Sprint(args...))
	}
	log.Panic(args...)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	if log == nil {
		panic(fmt.Sprintf(format, args...))
	}
	log.Panicf(format, args...)
}

// WithField creates an entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	if log == nil {
		return logrus.NewEntry(logrus.New()).WithField(key, value)
	}
	return log.WithField(key, value)
}

// WithFields creates an entry with multiple fields
func WithFields(fields map[string]interface{}) *logrus.Entry {
	if log == nil {
		return logrus.NewEntry(logrus.New()).WithFields(fields)
	}
	return log.WithFields(fields)
}

// WithError creates an entry with an error field
func WithError(err error) *logrus.Entry {
	if log == nil {
		return logrus.NewEntry(logrus.New()).WithError(err)
	}
	return log.WithError(err)
}

// GetLogger returns the underlying logrus logger
func GetLogger() *logrus.Logger {
	if log == nil {
		Init("info", "text")
	}
	return log
}