package logging

import (
	"bank-api/internal/config"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "INFO"
	}
}

type Logger struct {
	level  Level
	format string
	logger *log.Logger
}

type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

var defaultLogger *Logger

func Init(cfg *config.Config) {
	level := parseLevel(cfg.Logging.Level)
	defaultLogger = &Logger{
		level:  level,
		format: cfg.Logging.Format,
		logger: log.New(os.Stdout, "", 0),
	}
}

func parseLevel(levelStr string) Level {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) log(level Level, message string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
		Fields:    fields,
	}

	var output string
	if l.format == "json" {
		jsonData, _ := json.Marshal(entry)
		output = string(jsonData)
	} else {
		output = fmt.Sprintf("[%s] %s %s", entry.Timestamp, entry.Level, entry.Message)
		if len(fields) > 0 {
			fieldsStr, _ := json.Marshal(fields)
			output += fmt.Sprintf(" %s", fieldsStr)
		}
	}

	l.logger.Println(output)
}

func Debug(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		var f map[string]interface{}
		if len(fields) > 0 {
			f = fields[0]
		}
		defaultLogger.log(DEBUG, message, f)
	}
}

func Info(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		var f map[string]interface{}
		if len(fields) > 0 {
			f = fields[0]
		}
		defaultLogger.log(INFO, message, f)
	}
}

func Warn(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		var f map[string]interface{}
		if len(fields) > 0 {
			f = fields[0]
		}
		defaultLogger.log(WARN, message, f)
	}
}

func Error(message string, err error, fields map[string]interface{}) {
	if defaultLogger != nil {
		if fields == nil {
			fields = make(map[string]interface{})
		}
		if err != nil {
			fields["error"] = err.Error()
		}
		defaultLogger.log(ERROR, message, fields)
	}
}
