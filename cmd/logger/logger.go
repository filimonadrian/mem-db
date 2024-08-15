package logger

type LogLevel int

const LoggerKey string = "logger"
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
}

type LoggerOptions struct {
	LogLevel    string `json:"logLevel"`
	LogFilepath string `json:"logFilepath"`
	Console     bool   `json:"console`
}
