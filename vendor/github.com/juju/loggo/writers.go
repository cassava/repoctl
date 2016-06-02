// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"fmt"
	"sync"
	"time"
)

// Writers holds a set of Writers and provides operations for
// acting on that set. It also acts as a single Writer.
type Writers struct {
	mu               sync.Mutex
	combinedMinLevel Level
	all              map[string]MinLevelWriter
	init             func() map[string]MinLevelWriter
}

// NewWriters creates a new set of Writers using the provided
// details.
func NewWriters(initial map[string]MinLevelWriter) *Writers {
	ws := &Writers{}
	ws.reset(initial)
	return ws
}

// reset puts the list of Writers back into the initial state.
func (ws *Writers) reset(initial map[string]MinLevelWriter) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.all = make(map[string]MinLevelWriter)
	for name, writer := range initial {
		ws.addUnlocked(name, writer)
	}
	ws.resetMinLevel()
}

// AddWithLevel adds the writer to the list of Writers that get notified
// when Write() is called. When adding, the caller specifies the
// minimum logging level that will be written and a name for the
// writer. The name is used to identify the writer later (e.g. when
// removing it).
//
// If there is already a writer with that name, an error is returned.
func (ws *Writers) AddWithLevel(name string, writer Writer, minLevel Level) error {
	if writer == nil {
		return fmt.Errorf("Writer cannot be nil")
	}
	return ws.add(name, NewMinLevelWriter(writer, minLevel))
}

// add adds the writer to the list of Writers that get notified when
// Write() is called. When adding, the caller specifies the name for
// the writer. The name is used to identify the writer later (e.g. when
// removing it).
//
// If there is already a writer with that name, an error is returned.
func (ws *Writers) add(name string, writer MinLevelWriter) error {
	if writer == nil {
		return fmt.Errorf("Writer cannot be nil")
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if err := ws.addUnlocked(name, writer); err != nil {
		return err
	}

	ws.resetMinLevel()
	return nil
}

func (ws *Writers) addUnlocked(name string, writer MinLevelWriter) error {
	if _, found := ws.all[name]; found {
		return fmt.Errorf("there is already a Writer with the name %q", name)
	}
	ws.all[name] = writer
	return nil
}

// remove drops the Writer identified by 'name' from the set and
// returns that writer. If the Writer is not found, an error is
// returned.
func (ws *Writers) remove(name string) (MinLevelWriter, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	writer, err := ws.removeUnlocked(name)
	if err != nil {
		return nil, err
	}

	ws.resetMinLevel()
	return writer, nil
}

func (ws *Writers) removeUnlocked(name string) (MinLevelWriter, error) {
	writer, found := ws.all[name]
	if !found {
		return nil, fmt.Errorf("Writer %q is not recognized", name)
	}
	delete(ws.all, name)
	return writer, nil
}

// replace is a convenience method that does the atomic equivalent of
// calling remove() and then add(). The previous writer, which
// must exist, is returned.
func (ws *Writers) replace(name string, newWriter Writer) (Writer, error) {
	if newWriter == nil {
		return nil, fmt.Errorf("Writer cannot be nil")
	}
	ws.mu.Lock()
	defer ws.mu.Unlock()

	oldWriter, err := ws.removeUnlocked(name)
	if err != nil {
		return nil, err
	}
	replacement := NewMinLevelWriter(newWriter, oldWriter.MinLogLevel())
	if err := ws.addUnlocked(name, replacement); err != nil {
		return nil, err
	}
	ws.resetMinLevel()
	return oldWriter, nil
}

// MinLogLevel returns the minimum log level at which at least one of
// the writers will write.
func (ws *Writers) MinLogLevel() Level {
	return ws.combinedMinLevel
}

// Write implements Writer, sending the message to each known writer.
func (ws *Writers) Write(level Level, module, filename string, line int, timestamp time.Time, message string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for _, writer := range ws.all {
		if IsLevelEnabled(writer, level) {
			writer.Write(level, module, filename, line, timestamp, message)
		}
	}
}

func (ws *Writers) resetMinLevel() {
	// We assume the lock is already held
	combinedLevel := UNSPECIFIED
	if len(ws.all) > 0 {
		combinedLevel = CRITICAL
		for _, writer := range ws.all {
			minLevel := writer.MinLogLevel()
			if minLevel < combinedLevel {
				combinedLevel = minLevel
			}
		}
	}
	ws.combinedMinLevel.set(combinedLevel)
}
