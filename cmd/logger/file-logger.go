package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type FileLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
	level LogLevel
	file  *os.File
}

func NewFileLogger(options *LoggerOptions) (Logger, error) {

	file, err := os.OpenFile(options.LogFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("Could not open or create log file: %v", err)
	}

	logger := &FileLogger{
		debug: log.New(file, "DEBUG: ", log.LstdFlags),
		info:  log.New(file, "INFO: ", log.LstdFlags),
		warn:  log.New(file, "WARN: ", log.LstdFlags),
		error: log.New(file, "ERROR: ", log.LstdFlags),
		file:  file,
	}

	logger.SetLevel(options.LogLevel)
	return logger, nil
}

func (l *FileLogger) SetLevel(level string) {
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

func (l *FileLogger) Debug(v ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Println(v...)
	}
}

func (l *FileLogger) Info(v ...interface{}) {
	if l.level <= INFO {
		l.info.Println(v...)
	}
}

func (l *FileLogger) Warn(v ...interface{}) {
	if l.level <= WARN {
		l.warn.Println(v...)
	}
}

func (l *FileLogger) Error(v ...interface{}) {
	if l.level <= ERROR {
		l.error.Println(v...)
	}
}

func (l *FileLogger) Close() error {
	return l.file.Close()
}
