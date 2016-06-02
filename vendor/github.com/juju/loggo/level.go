// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"strings"
	"sync/atomic"
)

// The severity levels. Higher values are more considered more
// important.
const (
	UNSPECIFIED Level = iota
	TRACE
	DEBUG
	INFO
	WARNING
	ERROR
	CRITICAL
)

// Level holds a severity level.
type Level uint32

// ParseLevel converts a string representation of a logging level to a
// Level. It returns the level and whether it was valid or not.
func ParseLevel(level string) (Level, bool) {
	level = strings.ToUpper(level)
	switch level {
	case "UNSPECIFIED":
		return UNSPECIFIED, true
	case "TRACE":
		return TRACE, true
	case "DEBUG":
		return DEBUG, true
	case "INFO":
		return INFO, true
	case "WARN", "WARNING":
		return WARNING, true
	case "ERROR":
		return ERROR, true
	case "CRITICAL":
		return CRITICAL, true
	}
	return UNSPECIFIED, false
}

// String implements Stringer.
func (level Level) String() string {
	switch level {
	case UNSPECIFIED:
		return "UNSPECIFIED"
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	case CRITICAL:
		return "CRITICAL"
	}
	return "<unknown>"
}

// get atomically gets the value of the given level.
func (level *Level) get() Level {
	return Level(atomic.LoadUint32((*uint32)(level)))
}

// set atomically sets the value of the receiver
// to the given level.
func (level *Level) set(newLevel Level) {
	atomic.StoreUint32((*uint32)(level), uint32(newLevel))
}

// HasMinLevel represents values that have a minimum log level.
type HasMinLevel interface {
	// MinLogLevel returns the configured minimum log level of the
	// value. This is the level at which messages with a lower level
	// will be discarded.
	MinLogLevel() Level
}

// HasParentWithLevel represents values that have a parent that in turn
// have a minimum log level.
type HasParentWithMinLevel interface {
	// ParentWithMinLogLevel returns the value's parent (or nil).
	ParentWithMinLogLevel() HasMinLevel
}

// EffectiveLogMinLevel returns the effective minimum log level of the
// leveler. This is the level at which messages with a lower level
// will be discarded for this leveler.
//
// If the leveler returns a level of UNSPECIFIED (i.e. was configured
// without a log level) then the effective log level of the leveler's
// parent (if any) is returned.
func EffectiveMinLevel(leveler HasMinLevel) Level {
	if leveler == nil {
		// Under normal circumstances, there will always be a root
		// module with a non-UNSPECIFIED level.
		return UNSPECIFIED
	}
	// Perhaps check for an EffectiveMinLogLevel method right here...
	level := leveler.MinLogLevel()
	if level == UNSPECIFIED {
		// Get the level from the parent, if there is one.
		leveler, ok := leveler.(HasParentWithMinLevel)
		if ok {
			if parent := leveler.ParentWithMinLogLevel(); parent != nil {
				// We might consider guarding against cycles here...
				level = EffectiveMinLevel(parent)
			}
		}
	}
	return level
}

// IsLevelEnabled returns whether or not the leveler will honor log
// records at or above the given log level. The effective log level
// is used, meaning if the leveler does not specify a level then the
// level of its parent (if any) is used.
func IsLevelEnabled(leveler HasMinLevel, level Level) bool {
	minLevel := EffectiveMinLevel(leveler)
	return level >= minLevel
}
