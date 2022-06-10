package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type any = interface{}

type Level int

const (
	Verbose Level = iota
	Debug
	Info
	Warning
	Error
	Fatal
)

const (
	strVerbose = "Verbose"
	strDebug   = "Debug  "
	strInfo    = "Info   "
	strWarning = "Warning"
	strError   = "Error  "
	strFatal   = "Fatal  "
)

type Logger struct {
	tag       string
	level     Level
	timestamp bool
}

func NewLogger(tag string) Logger {
	// set log level to Debug by default
	return Logger{tag: tag, level: Debug, timestamp: true}
}

func (log *Logger) SetTag(tag string) {
	log.tag = tag
}

func (log *Logger) SetLevel(level Level) {
	log.level = level
}

func (log *Logger) GetLevel() Level {
	return log.level
}

func (log *Logger) ShowTimestamp(show bool) {
	log.timestamp = show
}

func (log *Logger) Print(level Level, f string, msg ...any) {
	if log.level <= level {
		switch level {
		case Verbose:
			color.Set(color.FgCyan)
		case Debug:
			color.Set(color.FgWhite)
		case Info:
			color.Set(color.FgGreen)
		case Warning:
			color.Set(color.FgYellow)
		case Error:
			color.Set(color.FgRed)
		case Fatal:
			color.Set(color.FgHiMagenta)
		}

		fmt.Printf(f, msg...)

		color.Unset()
	}
}

func (log *Logger) LogFormat(level Level, f string, msg ...any) {
	if log.level <= level {
		var levelText string
		var timeText string

		if log.timestamp {
			timeText = time.Now().Local().Format("2006-01-02 15:04:05.06")
		}

		switch level {
		case Verbose:
			color.Set(color.FgCyan)
			levelText = strVerbose
		case Debug:
			color.Set(color.FgWhite)
			levelText = strDebug
		case Info:
			color.Set(color.FgGreen)
			levelText = strInfo
		case Warning:
			color.Set(color.FgYellow)
			levelText = strWarning
		case Error:
			color.Set(color.FgRed)
			levelText = strError
		case Fatal:
			color.Set(color.FgHiMagenta)
			levelText = strFatal
		}

		if log.timestamp {
			fmt.Printf("%s) %s [%s] ", timeText, levelText, log.tag)
		} else {
			fmt.Printf("%s [%s] ", levelText, log.tag)
		}
		fmt.Printf(f, msg...)
		fmt.Println()

		color.Unset()
	}
}

func (log *Logger) Log(level Level, msg ...any) {
	if log.level <= level {
		var levelText string
		var timeText string

		if log.timestamp {
			timeText = time.Now().Local().Format("2006-01-02 15:04:05.06")
		}

		switch level {
		case Verbose:
			color.Set(color.FgCyan)
			levelText = strVerbose
		case Debug:
			color.Set(color.FgWhite)
			levelText = strDebug
		case Info:
			color.Set(color.FgGreen)
			levelText = strInfo
		case Warning:
			color.Set(color.FgYellow)
			levelText = strWarning
		case Error:
			color.Set(color.FgRed)
			levelText = strError
		case Fatal:
			color.Set(color.FgHiMagenta)
			levelText = strFatal
		}

		if log.timestamp {
			fmt.Printf("%s) %s [%s] ", timeText, levelText, log.tag)
		} else {
			fmt.Printf("%s [%s] ", levelText, log.tag)
		}
		fmt.Printf(strings.Repeat("%v", len(msg)), msg...)
		fmt.Println()

		color.Unset()
	}
}

func (log *Logger) LogV(msg ...any) {
	log.Log(Verbose, msg...)
}

func (log *Logger) LogD(msg ...any) {
	log.Log(Debug, msg...)
}

func (log *Logger) LogI(msg ...any) {
	log.Log(Info, msg...)
}

func (log *Logger) LogW(msg ...any) {
	log.Log(Warning, msg...)
}

func (log *Logger) LogE(msg ...any) {
	log.Log(Error, msg...)
}

func (log *Logger) LogF(msg ...any) {
	log.Log(Fatal, msg...)
}
