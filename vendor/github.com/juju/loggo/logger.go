// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// A Logger represents a logging module. It has an associated logging
// level which can be changed; messages of lesser severity will
// be dropped. Loggers have a hierarchical relationship - see
// the package documentation.
//
// The zero Logger value is usable - any messages logged
// to it will be sent to the root Logger.
type Logger struct {
	impl   *module
	writer MinLevelWriter
}

// NewRootLogger creates a root logger and returns it, along with
// the writers that the logger will use.
func NewRootLogger() (Logger, *Writers) {
	writers := NewWriters(nil) // starts off empty
	logger := Logger{
		impl:   newRootModule(),
		writer: writers,
	}
	return logger, writers
}

// NewLogger creates a new Logger with the given name and parent.
// The logger uses the name to identify itself. The parent is used when
// determining the effective log level. The new logger is returned,
// along with the writers the logger will use.
func NewLogger(name string, parent Logger) (Logger, *Writers) {
	var writers *Writers
	if parent.isZero() {
		parent, writers = NewRootLogger()
	} else {
		writers = NewWriters(nil) // starts off empty
		if parent.writer != nil {
			// We set the level as low as possible in order to defer
			// strictly to the new logger's level.
			// TODO(ericsnow) Use parent.writer's level.
			writers.AddWithLevel(defaultWriterName, parent.writer, UNSPECIFIED)
		}
	}
	logger := newLogger(name, parent.getModule(), writers)
	return logger, writers
}

func newLogger(name string, parent *module, writer MinLevelWriter) Logger {
	// The parent *may* be nil.
	name = strings.ToLower(name)
	return Logger{
		impl: &module{
			name:   name,
			level:  UNSPECIFIED,
			parent: parent,
		},
		writer: writer,
	}
}

func (logger Logger) isZero() bool {
	return reflect.DeepEqual(logger, Logger{})
}

func (logger Logger) getModule() *module {
	if logger.impl == nil {
		// This is a sensible root module to use for zero values.
		return newRootModule()
	}
	return logger.impl
}

// Name returns the logger's module name.
func (logger Logger) Name() string {
	return logger.getModule().Name()
}

// LogLevel returns the configured min log level of the logger.
func (logger Logger) LogLevel() Level {
	return logger.getModule().MinLogLevel()
}

// EffectiveLogLevel returns the effective min log level of
// the receiver - that is, messages with a lesser severity
// level will be discarded.
//
// If the log level of the receiver is unspecified,
// it will be taken from the effective log level of its
// parent.
func (logger Logger) EffectiveLogLevel() Level {
	return EffectiveMinLevel(logger.getModule())
}

// Config returns the current configuration for the Logger.
func (logger Logger) Config() LoggerConfig {
	cfg := logger.getModule().config()
	logger.updateConfig(&cfg)
	return cfg
}

// updateConfig adds any logger-specific info to the provided config.
func (logger Logger) updateConfig(cfg *LoggerConfig) {
	// For now there isn't any logger-specific info.
}

// ApplyConfig configures the logger according to the provided config.
func (logger Logger) ApplyConfig(cfg LoggerConfig) {
	logger.getModule().applyConfig(cfg)
}

// SetLogLevel sets the severity level of the given logger.
// The root logger cannot be set to UNSPECIFIED level.
// See EffectiveLogLevel for how this affects the
// actual messages logged.
func (logger Logger) SetLogLevel(level Level) {
	logger.getModule().setLevel(level)
}

// Logf logs a printf-formatted message at the given level.
// A message will be discarded if level is less than the
// the effective log level of the logger.
// Note that the writers may also filter out messages that
// are less than their registered minimum severity level.
func (logger Logger) Logf(level Level, message string, args ...interface{}) {
	logger.LogCallf(2, level, message, args...)
}

// LogCallf logs a printf-formatted message at the given level.
// The location of the call is indicated by the calldepth argument.
// A calldepth of 1 means the function that called this function.
// A message will be discarded if level is less than the
// the effective log level of the logger.
// Note that the writers may also filter out messages that
// are less than their registered minimum severity level.
func (logger Logger) LogCallf(calldepth int, level Level, message string, args ...interface{}) {
	if !logger.willWrite(level) {
		return
	}
	// Gather time, and filename, line number.
	now := time.Now() // get this early.
	// Param to Caller is the call depth.  Since this method is called from
	// the Logger methods, we want the place that those were called from.
	_, file, line, ok := runtime.Caller(calldepth + 1)
	if !ok {
		file = "???"
		line = 0
	}
	// Trim newline off format string, following usual
	// Go logging conventions.
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[0 : len(message)-1]
	}

	// To avoid having a proliferation of Info/Infof methods,
	// only use Sprintf if there are any args, and rely on the
	// `go vet` tool for the obvious cases where someone has forgotten
	// to provide an arg.
	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	}
	logger.writer.Write(level, logger.impl.name, file, line, now, formattedMessage)
}

func (logger Logger) willWrite(level Level) bool {
	if !IsLevelEnabled(logger.getModule(), level) {
		return false
	}
	if !IsLevelEnabled(logger.writer, level) {
		return false
	}
	if level < TRACE || level > CRITICAL {
		return false
	}
	return true
}

// Criticalf logs the printf-formatted message at critical level.
func (logger Logger) Criticalf(message string, args ...interface{}) {
	logger.Logf(CRITICAL, message, args...)
}

// Errorf logs the printf-formatted message at error level.
func (logger Logger) Errorf(message string, args ...interface{}) {
	logger.Logf(ERROR, message, args...)
}

// Warningf logs the printf-formatted message at warning level.
func (logger Logger) Warningf(message string, args ...interface{}) {
	logger.Logf(WARNING, message, args...)
}

// Infof logs the printf-formatted message at info level.
func (logger Logger) Infof(message string, args ...interface{}) {
	logger.Logf(INFO, message, args...)
}

// Debugf logs the printf-formatted message at debug level.
func (logger Logger) Debugf(message string, args ...interface{}) {
	logger.Logf(DEBUG, message, args...)
}

// Tracef logs the printf-formatted message at trace level.
func (logger Logger) Tracef(message string, args ...interface{}) {
	logger.Logf(TRACE, message, args...)
}

// IsLevelEnabled returns whether debugging is enabled
// for the given log level.
func (logger Logger) IsLevelEnabled(level Level) bool {
	return IsLevelEnabled(logger.getModule(), level)
}

// IsErrorEnabled returns whether debugging is enabled
// at error level.
func (logger Logger) IsErrorEnabled() bool {
	return logger.IsLevelEnabled(ERROR)
}

// IsWarningEnabled returns whether debugging is enabled
// at warning level.
func (logger Logger) IsWarningEnabled() bool {
	return logger.IsLevelEnabled(WARNING)
}

// IsInfoEnabled returns whether debugging is enabled
// at info level.
func (logger Logger) IsInfoEnabled() bool {
	return logger.IsLevelEnabled(INFO)
}

// IsDebugEnabled returns whether debugging is enabled
// at debug level.
func (logger Logger) IsDebugEnabled() bool {
	return logger.IsLevelEnabled(DEBUG)
}

// IsTraceEnabled returns whether debugging is enabled
// at trace level.
func (logger Logger) IsTraceEnabled() bool {
	return logger.IsLevelEnabled(TRACE)
}
