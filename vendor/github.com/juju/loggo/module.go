// Copyright 2016 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package loggo

import (
	"strings"
	"sync"
)

// Do not change rootName: modules.resolve() will misbehave if it isn't "".
const (
	rootName         = ""
	rootString       = "<root>"
	defaultRootLevel = WARNING
	defaultLevel     = UNSPECIFIED
)

type module struct {
	name   string
	level  Level
	parent *module
}

func newModule(name string, parent *module) *module {
	if name == rootString {
		name = rootName
	}
	if name == rootName {
		if parent != nil {
			panic("should never happen")
		}
		return newRootModule()
	}
	return newSubmodule(name, parent, defaultLevel)
}

func newRootModule() *module {
	return &module{
		name:  rootName,
		level: defaultRootLevel,
	}
}

func newSubmodule(name string, parent *module, level Level) *module {
	if parent == nil {
		// We must ensure that every non-root module has a root ancestor.
		parent = newRootModule()
	}
	return &module{
		name:   name,
		level:  level,
		parent: parent,
	}
}

// Name returns the module's name.
func (module *module) Name() string {
	if module.name == rootName {
		return "<root>"
	}
	return module.name
}

// MinLogLevel returns the configured minimum log level of the
// module. This is the level at which messages with a lower level
// will be discarded.
func (module *module) MinLogLevel() Level {
	return module.level.get()
}

// ParentWithMinLogLevel returns the module's parent (or nil).
func (module *module) ParentWithMinLogLevel() HasMinLevel {
	if module.parent == nil { // avoid double nil
		return nil
	}
	return module.parent
}

// config returns the current configuration for the module.
func (module *module) config() LoggerConfig {
	return LoggerConfig{
		Level: module.MinLogLevel(),
	}
}

// applyConfig configures the logger according to the provided config.
func (module *module) applyConfig(cfg LoggerConfig) {
	module.setLevel(cfg.Level)
}

// setLevel sets the severity level of the given module.
// The root module cannot be set to UNSPECIFIED level.
func (module *module) setLevel(level Level) {
	// The root module can't be unspecified (see Logger.EffectiveLogLevel).
	if module.name == rootName && level == UNSPECIFIED {
		level = defaultRootLevel
	}
	module.level.set(level)
}

type modules struct {
	mu           sync.Mutex
	rootLevel    Level
	defaultLevel Level
	all          map[string]*module
}

// Initially the modules map only contains the root module.
func newModules(rootLevel Level) *modules {
	m := &modules{
		rootLevel:    rootLevel,
		defaultLevel: defaultLevel,
	}
	m.initUnlocked()
	return m
}

func (m *modules) initUnlocked() {
	if m.rootLevel <= UNSPECIFIED {
		// The root level cannot be UNSPECIFIED.
		m.rootLevel = defaultRootLevel
	}
	if m.defaultLevel <= UNSPECIFIED {
		m.defaultLevel = defaultLevel
	}
	root := newRootModule()
	root.level = m.rootLevel
	m.all = map[string]*module{
		rootName: root,
	}
}

func (m *modules) maybeInitUnlocked() {
	if m.all == nil {
		m.initUnlocked()
	}
}

// get returns a Logger for the given module name,
// creating it and its parents if necessary.
func (m *modules) get(name string) *module {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maybeInitUnlocked() // guarantee we have a root module

	// Lowercase the module name, and look for it in the modules map.
	name = strings.ToLower(name)
	return m.resolveUnlocked(name)
}

func (m *modules) resolveUnlocked(name string) *module {
	// m must already be initialized (e.g. newModules()).
	if name == rootString {
		name = rootName
	}
	if impl, found := m.all[name]; found {
		return impl
	}
	parentName := rootName
	if i := strings.LastIndex(name, "."); i >= 0 {
		parentName = name[0:i]
	}
	// Since there is always a root module, we always get a parent here.
	parent := m.resolveUnlocked(parentName)
	impl := newSubmodule(name, parent, m.defaultLevel)
	m.all[name] = impl
	return impl
}

// config returns the current configuration of the modules. Modules
// with UNSPECIFIED level will not be included.
func (m *modules) config() LoggersConfig {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg := make(LoggersConfig)
	for _, module := range m.all {
		if module.MinLogLevel() <= UNSPECIFIED {
			continue
		}
		cfg[module.Name()] = module.config()
	}
	return cfg
}

// resetLevels iterates through the known modules and sets the levels of all
// to UNSPECIFIED, except for <root> which is set to WARNING.
func (m *modules) resetLevels() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, module := range m.all {
		if name == rootName {
			module.level.set(m.rootLevel)
		} else {
			module.level.set(m.defaultLevel)
		}
	}
}
