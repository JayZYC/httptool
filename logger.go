package httptool

import (
	"context"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// Colors
const (
	Reset    = "\033[0m"
	Red      = "\033[31m"
	Green    = "\033[32m"
	Yellow   = "\033[33m"
	Magenta  = "\033[35m"
	BlueBold = "\033[34;1m"
)

// LogLevel log level
type LogLevel int

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
	// Debug debug log level
	Debug
)

// Writer log writer interface
type Writer interface {
	Printf(string, ...interface{})
}

// Config logger config
type Config struct {
	Colorful bool
	LogLevel LogLevel
}

// Interface logger interface
type Interface interface {
	LogMode(LogLevel) Interface
	Debug(context.Context, string, ...interface{})
	Info(context.Context, string, ...interface{})
	Warn(context.Context, string, ...interface{})
	Error(context.Context, string, ...interface{})
}

var (
	// Default logger
	Default = New(log.New(os.Stdout, "\r\n", log.LstdFlags), Config{
		LogLevel: Warn,
		Colorful: true,
	})
)

// New initialize logger
func New(writer Writer, config Config) Interface {
	var (
		debugStr = "%s\n[debug] "
		infoStr  = "%s\n[info] "
		warnStr  = "%s\n[warn] "
		errStr   = "%s\n[error] "
	)

	if config.Colorful {
		debugStr = Green + "%s\n" + Reset + Yellow + "[debug] " + Reset
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
	}

	return &logger{
		Writer:   writer,
		Config:   config,
		debugStr: debugStr,
		infoStr:  infoStr,
		warnStr:  warnStr,
		errStr:   errStr,
	}
}

type logger struct {
	Writer
	Config
	debugStr, infoStr, warnStr, errStr  string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *logger) LogMode(level LogLevel) Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Debug print debug messages
func (l *logger) Debug(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Debug {
		l.Printf(l.debugStr+msg, append([]interface{}{getLoggerCallerInfo()}, data...)...)
	}
}

// Info print info
func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Info {
		l.Printf(l.infoStr+msg, append([]interface{}{getLoggerCallerInfo()}, data...)...)
	}
}

// Warn print warn messages
func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{getLoggerCallerInfo()}, data...)...)
	}
}

// Error print error messages
func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= Error {
		l.Printf(l.errStr+msg, append([]interface{}{getLoggerCallerInfo()}, data...)...)
	}
}

// getLoggerCallerInfo 日志调用者信息 -- 文件名, 行号
func getLoggerCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	file = path.Base(file)
	return strings.Join([]string{file, strconv.Itoa(line)}, ":")
}
