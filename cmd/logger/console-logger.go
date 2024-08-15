package logger

import (
	"log"
	"os"
	"strings"
)

type ConsoleLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	level LogLevel
}

func NewConsoleLogger(options *LoggerOptions) (Logger, error) {
	logger := &ConsoleLogger{
		debug: log.New(os.Stdout, "DEBUG: ", log.LstdFlags),
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}

	logger.SetLevel(options.LogLevel)
	return logger, nil
}

// func (l *ConsoleLogger) SetLevel(level LogLevel) {
// 	l.level = level
// }

func (l *ConsoleLogger) SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		l.level = DEBUG
	case "info":
		l.level = INFO
	case "warn":
		l.level = WARN
	case "error":
		l.level = ERROR
	default:
		l.level = INFO
	}
}

func (l *ConsoleLogger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Println(v...)
	}
}

func (l *ConsoleLogger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.info.Println(v...)
	}
}

func (l *ConsoleLogger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.warn.Println(v...)
	}
}

func (l *ConsoleLogger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.error.Println(v...)
	}
}
