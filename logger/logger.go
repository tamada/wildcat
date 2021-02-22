package logger

import (
	"log"
	"strings"
)

// Level shows the level for logging.
type Level int

const (
	// DEBUG shows the debug level.
	DEBUG Level = iota + 1
	// INFO shows the info level.
	INFO
	// WARN shows the warn level.
	WARN
	// FATAL shows the fatal level.
	FATAL
)

func LogLevel(level string) Level {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "fatal":
		return FATAL
	default:
		Warnf(`%s: unknown log level, use WARN`, level)
		return WARN
	}
}

func (level Level) String() string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is a struct for logging the message.
type Logger struct {
	outputLevel Level
}

var defaultLogger = New(WARN)

// GetLevel returns the log level of the default logger.
func GetLevel() Level {
	return defaultLogger.GetLevel()
}

// SetLevel set the log level for the default logger.
func SetLevel(Level Level) {
	defaultLogger.SetLevel(Level)
}

// Debug prints the given message when the log level of the default logger is DEBUG.
func Debug(message string) {
	defaultLogger.Debug(message)
}

// Info prints the given message when the log level of the default logger is less than INFO.
func Info(message string) {
	defaultLogger.Info(message)
}

// Warn prints the given message when the log level of the default logger is less than WARN.
func Warn(message string) {
	defaultLogger.Warn(message)
}

// Fatal prints the given message.
func Fatal(message string) {
	defaultLogger.Fatal(message)
}

// Debugf prints the given message by Printf format when the log level of the default logger is DEBUG.
func Debugf(format string, v ...interface{}) {
	defaultLogger.Debugf(format, v...)
}

// Infof prints the given message by Printf format when the log level of the default logger is INFO.
func Infof(format string, v ...interface{}) {
	defaultLogger.Infof(format, v...)
}

// Warnf prints the given message by Printf format when the log level of the default logger is WARN.
func Warnf(format string, v ...interface{}) {
	defaultLogger.Warnf(format, v...)
}

// Fatalf prints the given message by Printf format.
func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

// New creates the new logger with given log level and returns it.
func New(giveLevel Level) *Logger {
	return &Logger{outputLevel: giveLevel}
}

// Debug prints the given message when the log level of the receiver logger is DEBUG.
func (logger *Logger) Debug(message string) {
	if logger.outputLevel <= DEBUG {
		log.Println(message)
	}
}

// Info prints the given message when the log level of the receiver logger is INFO.
func (logger *Logger) Info(message string) {
	if logger.outputLevel <= INFO {
		log.Println(message)
	}
}

// Warn prints the given message when the log level of the receiver logger is WARN.
func (logger *Logger) Warn(message string) {
	if logger.outputLevel <= WARN {
		log.Println(message)
	}
}

// Fatal prints the given message when the log level of the receiver logger is FATAL.
func (logger *Logger) Fatal(message string) {
	if logger.outputLevel <= FATAL {
		log.Println(message)
	}
}

// Debugf prints the given message by Printf format when the log level of the receiver logger is DEBUG.
func (logger *Logger) Debugf(format string, v ...interface{}) {
	if logger.outputLevel <= DEBUG {
		log.Printf(format, v...)
	}
}

// Infof prints the given message by Printf format when the log level of the receiver logger is INFO.
func (logger *Logger) Infof(format string, v ...interface{}) {
	if logger.outputLevel <= INFO {
		log.Printf(format, v...)
	}
}

// Warnf prints the given message by Printf format when the log level of the receiver logger is WARN.
func (logger *Logger) Warnf(format string, v ...interface{}) {
	if logger.outputLevel <= WARN {
		log.Printf(format, v...)
	}
}

// Fatalf prints the given message by Printf format.
func (logger *Logger) Fatalf(format string, v ...interface{}) {
	if logger.outputLevel <= FATAL {
		log.Printf(format, v...)
	}
}

// GetLevel returns the log level of the receiver logger.
func (logger *Logger) GetLevel() Level {
	return logger.outputLevel
}

// SetLevel sets the log level by the given level for the receiver logger.
func (logger *Logger) SetLevel(giveLevel Level) {
	if giveLevel >= DEBUG && giveLevel <= FATAL {
		logger.outputLevel = giveLevel
	}
}
