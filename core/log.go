package core

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
)

// Variables for log types
const (
	FATAL     = 5
	ERROR     = 4
	WARN      = 3
	IMPORTANT = 2
	INFO      = 1
	DEBUG     = 0
)

// LogColors for logging types
var LogColors = map[int]*color.Color{
	FATAL:     color.New(color.FgRed).Add(color.Bold),
	ERROR:     color.New(color.FgRed),
	WARN:      color.New(color.FgYellow),
	IMPORTANT: color.New(color.Bold),
	DEBUG:     color.New(color.FgCyan).Add(color.Faint),
}

// Logger struct
type Logger struct {
	sync.Mutex
	debug  bool
	silent bool
}

// SetSilent to true or false
func (l *Logger) SetSilent(s bool) {
	l.silent = s
}

// SetDebug to true or false
func (l *Logger) SetDebug(d bool) {
	l.debug = d
}

// Log main function
func (l *Logger) Log(level int, format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	if level == DEBUG && !l.debug {
		return
	} else if level < ERROR && l.silent {
		return
	}

	if c, ok := LogColors[level]; ok {
		c.Printf(format, args...)
	} else {
		fmt.Printf(format, args...)
	}

	if level == FATAL {
		os.Exit(1)
	}
}

// Fatal to log and exit with code 1
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.Log(FATAL, format, args...)
}

// Error to log in red color
func (l *Logger) Error(format string, args ...interface{}) {
	l.Log(ERROR, format, args...)
}

// Warn to log in yellow color
func (l *Logger) Warn(format string, args ...interface{}) {
	l.Log(WARN, format, args...)
}

// Important to log message in bold
func (l *Logger) Important(format string, args ...interface{}) {
	l.Log(IMPORTANT, format, args...)
}

// Info to log regular message
func (l *Logger) Info(format string, args ...interface{}) {
	l.Log(INFO, format, args...)
}

// Debug to log in blue color
func (l *Logger) Debug(format string, args ...interface{}) {
	l.Log(DEBUG, format, args...)
}
