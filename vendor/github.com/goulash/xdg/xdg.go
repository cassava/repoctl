// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Package xdg provides an implementation of the XDG Base Directory Specification.
//
// On initialization, the following package variables are set to their recommended
// values, by reading the corresponding environment variables and falling back to
// specification defaults if necessary.
//
//     ConfigHome      // user configuration base directory, e.g. ~/.config
//     DataHome        // user data files base directory, e.g. ~/.local/share
//     CacheHome       // user cache files base directory, e.g. ~/.cache
//     RuntimeDir      // user runtime files base directory, e.g. /run/user/1000
//     ConfigDirs      // global configuration directories, e.g. /etc/xdg
//     DataDirs        // global data files directories, e.g. /usr/local/share
//     AllConfigDirs   // user and global configuration directories
//     AllDataDirs     // user and global data directories
//
// Initialization happens automatically, but can also be explicitely started with
// the Init function. If no valid path can be constructed, the variable is left
// blank or empty. If one of the required paths is blank or empty, the program
// should fail. These variables should be treated as read-only; change them only
// if you know what you are doing.
//
// The package has four classes of functions, which should suffice for most needs:
//
//     User*           // construct a valid path for user (config|data|...) files
//     Find*           // find existing (config|data|...) files
//     Merge*          // execute a function on each found (config|data) file
//     Open*           // open or create a user (config|data|...) file
//
// Only the Open* functions may alter the filesystem in any way: this is
// restricted to creating XDG user base directories and files therein. Directories
// in ConfigDirs and DataDirs are not modified.
//
// The XDG Base Directory Specification, henceforth “the specification”, defines
// several types of files: configuration, data, cache, and runtime files.
// The specification can be found at:
//
//     http://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html
//
// Configuration files
//
// Configuration files are read from ConfigHome and from ConfigDirs;
// they are only written in ConfigHome.
//
// ConfigHome is a single base directory relative to which user-specific data
// files should be written. This directory is defined by the environment
// variable $XDG_DATA_HOME. If $XDG_CONFIG_HOME is not set, the default
// "$HOME/.config" is used.
//
// ConfigDirs is a set of preference ordered base directories relative to
// which configuration files should be searched. This set of directories is
// defined by the environment variable $XDG_CONFIG_DIRS. The directories in
// $XDG_CONFIG_DIRS should be seperated with a colon ':'. If $XDG_CONFIG_DIRS
// is not set, the default "/etc/xdg" is used.
//
// ConfigHomeDirs combines ConfigHome and ConfigDirs into one preference
// ordered set of directories.
//
// Data files
//
// Data files are read from DataHome and from DataDirs;
// they are only written in DataHome.
//
// DataHome is a single base directory relative to which user-specific data files
// should be written. This directory is defined by the environment variable
// $XDG_DATA_HOME. If $XDG_DATA_HOME is not set, the default "$HOME/.local/share"
// is used.
//
// DataDirs is a set of preference ordered base directories relative to which data
// files should be searched. This set of directories is defined by the environment
// variable $XDG_DATA_DIRS. If $XDG_CONFIG_DIRS is not set, the default
// "/usr/local/share:/usr/share" is used.
//
// Cache files
//
// CacheHome is a single base directory relative to which user-specific
// non-essential (cached) data should be written. This directory is defined by the
// environment variable $XDG_CACHE_HOME.  If $XDG_CACHE_HOME is not set, the
// default "$HOME/.cache" is used.
//
// Runtime files
//
// RuntimeDir is a single base directory relative to which user-specific
// runtime files and other file objects should be placed. This directory is
// defined by the environment variable $XDG_RUNTIME_DIR. If $XDG_RUNTIME_DIR
// is not set, the following method is used to find an appropriate directory:
//
//     path.Join(os.TempDir(), fmt.Sprintf("xdg-%d", os.Getuid()))
//
// This usually results in paths such as "/tmp/xdg-1000". Normally, we expect
// something along the lines of "/run/user/1000".
//
// In this implementation, we assume that the system takes care of removing the
// XDG runtime directory at shutdown.
package xdg

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

// Getenv reads several environment variables. You can provide your own
// implementation if you have special needs (e.g. mock testing).
// If you change Getenv, you need to call Init() again.
// The following variables are read:
//
//  HOME
//  XDG_CONFIG_HOME
//  XDG_DATA_HOME
//  XDG_CACHE_HOME
//  XDG_RUNTIME_DIR
//  XDG_CONFIG_DIRS
//  XDG_DATA_DIRS
var Getenv func(string) string = os.Getenv

var (
	// Errors contains all errors that occurred during initialization.
	Errors []error

	// ErrInvalidHome is found in the Errors slice if the HOME environment variable
	// is not set or it is not an absolute path.
	ErrInvalidHome = errors.New("environment variable HOME is invalid or not set")

	// ErrInvalidPath is returned when attempting to create or open an invalid path.
	// This means that some XDG variable could not be correctly set.
	ErrInvalidPath = errors.New("invalid XDG path used")
)

var (
	// ConfigHome is a single base directory relative to which user-specific
	// configuration files should be written.
	ConfigHome string

	// DataHome is a single base directory relative to which user-specific data
	// files should be written.
	DataHome string

	// CacheHome is a single base directory relative to which user-specific
	// non-essential (cached) data should be written.
	CacheHome string

	// RuntimeDir is a single base directory relative to which user-specific
	// runtime files and other file objects should be placed.
	RuntimeDir string

	// ConfigDirs is a set of preference ordered base directories relative to
	// which configuration files should be searched.
	ConfigDirs []string

	// DataDirs is a set of preference ordered base directories relative to
	// which data files should be searched.
	DataDirs []string

	// ConfigHomeDirs is the same as ConfigDirs, with ConfigHome at first place.
	ConfigHomeDirs []string

	// DataHomeDirs is the same as DataDirs, with DataHome at first place.
	DataHomeDirs []string

	// home is a single base directory of the user's home directory.
	// This directory is defined by the environment variable $HOME.
	//
	// If $HOME is not set, and is required, then other variables might be empty.
	home string
)

func init() {
	Init()
}

// Init initializes this package, reading several environment variables
// (using Getenv, which you can override if you need to), and setting
// several package variables.
//
// It is normally not necessary to call Init; you only need to do so
// if you would like to reset the package (e.g. because you changed
// Getenv).
func Init() {
	Errors = []error{}
	home = Getenv("HOME")
	if !path.IsAbs(home) {
		home = ""
		Errors = append(Errors, ErrInvalidHome)
	}

	ConfigHome = xdgPath("XDG_CONFIG_HOME", "$HOME/.config")
	DataHome = xdgPath("XDG_DATA_HOME", "$HOME/.local/share")
	CacheHome = xdgPath("XDG_CACHE_HOME", "$HOME/.cache")
	tmp := path.Join(os.TempDir(), fmt.Sprintf("xdg-%d", os.Getuid()))
	RuntimeDir = xdgPath("XDG_RUNTIME_DIR", tmp)
	ConfigDirs = xdgPaths("XDG_CONFIG_DIRS", "/etc/xdg")
	DataDirs = xdgPaths("XDG_DATA_DIRS", "/usr/local/share:/usr/share")
	ConfigHomeDirs = combine(ConfigHome, ConfigDirs)
	DataHomeDirs = combine(DataHome, DataDirs)
}

func xdgPath(env, def string) string {
	x := Getenv(env)

	if x == "" {
		if strings.Contains(def, "$HOME") {
			if home != "" {
				x = strings.Replace(def, "$HOME", home, -1)
			}
		} else {
			x = def
		}
	}

	// The XDG specification states:
	//
	//  All paths set in these environment variables must be absolute. If an
	//  implementation encounters a relative path in any of these variables it
	//  should consider the path invalid and ignore it.
	if path.IsAbs(x) {
		return x
	}
	Errors = append(Errors, errors.New("no value set for "+env))
	return ""
}

func xdgPaths(env, def string) []string {
	xs := Getenv(env)

	if xs == "" {
		xs = def
	}

	var fs []string
	for _, x := range strings.Split(xs, string(os.PathListSeparator)) {
		// See comment in xdgPath.
		if path.IsAbs(x) {
			fs = append(fs, x)
		} else {
			Errors = append(Errors, errors.New("ignoring "+env+" path element: "+x))
		}
	}
	return fs
}

// combine x and xs to a single slice, where x is in the front.
// If x is empty, xs is returned.
func combine(x string, xs []string) []string {
	if x == "" {
		return xs
	}

	n := len(xs) + 1
	ns := make([]string, n)

	ns[0] = x
	for i := 1; i < n; i++ {
		ns[i] = xs[i-1]
	}
	return ns
}

func UserConfig(file string) string  { return join(ConfigHome, file) }
func UserData(file string) string    { return join(DataHome, file) }
func UserCache(file string) string   { return join(CacheHome, file) }
func UserRuntime(file string) string { return join(RuntimeDir, file) }

func join(dir, file string) string {
	if dir == "" {
		return ""
	}
	p := path.Join(dir, file)
	if !path.IsAbs(p) {
		return ""
	}
	return p
}

func FindConfig(file string) string      { return find(file, ConfigHomeDirs) }
func FindData(file string) string        { return find(file, DataHomeDirs) }
func FindCache(file string) string       { return find(file, []string{CacheHome}) }
func FindRuntime(file string) string     { return find(file, []string{RuntimeDir}) }
func FindAllConfig(file string) []string { return findAll(file, ConfigHomeDirs) }
func FindAllData(file string) []string   { return findAll(file, DataHomeDirs) }

// find returns the first file that exists, else "".
func find(file string, paths []string) string {
	for _, dir := range paths {
		p := join(dir, file)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		return p
	}
	return ""
}

func findAll(file string, paths []string) []string {
	ps := make([]string, 0, len(paths))
	for _, dir := range paths {
		p := join(dir, file)
		if _, err := os.Stat(p); err != nil {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}

// MergeFunc is given to the Merge* functions to handle the files that it
// finds. It receives an absolute path to a file, which MergeFunc can then try
// to open. When MergeFunc is done with the file (for example, it couldn't read
// the file, or it was empty) then it can return nil. If an error is returned,
// then the Merge* function aborts and returns this error. If an error
// hasn't occurred, but no files need be further inspected, Skip can be returned.
type MergeFunc func(filepath string) error

// Skip can be returned by a MergeFunc which causes the Merge* functions
// to skip the rest of the files to be merged.
var Skip = errors.New("skip the rest of the files to be merged")

func MergeConfig(file string, f MergeFunc) error  { return merge(file, f, ConfigHomeDirs) }
func MergeConfigR(file string, f MergeFunc) error { return mergeR(file, f, ConfigHomeDirs) }
func MergeData(file string, f MergeFunc) error    { return merge(file, f, DataHomeDirs) }
func MergeDataR(file string, f MergeFunc) error   { return mergeR(file, f, DataHomeDirs) }

func mergeR(file string, f MergeFunc, paths []string) error {
	var err error
	for s := range reverse(findAll(file, paths)) {
		if err = f(s); err != nil {
			break
		}
	}
	if err == Skip {
		return nil
	}
	return err
}

func merge(file string, f MergeFunc, paths []string) error {
	var err error
	for _, s := range findAll(file, paths) {
		if err = f(s); err != nil {
			break
		}
	}
	if err == Skip {
		return nil
	}
	return err
}

func reverse(xs []string) <-chan string {
	ch := make(chan string)
	go func() {
		for i := len(xs); i != 0; i-- {
			ch <- xs[i-1]
		}
		close(ch)
	}()
	return ch
}

func OpenConfig(file string, flag int) (*os.File, error) { return open(UserConfig(file), flag) }
func OpenData(file string, flag int) (*os.File, error)   { return open(UserData(file), flag) }
func OpenCache(file string, flag int) (*os.File, error)  { return open(UserCache(file), flag) }
func OpenRuntime(file string, flag int) (*os.File, error) {
	// TODO: Make sure that the runtime directory is only readable by the user.
	_, err := os.Stat(RuntimeDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(RuntimeDir, os.ModeDir|0700)
			if err != nil {
				return nil, err
			}
			_, err = os.Stat(RuntimeDir)
			if err != nil {
				// This really should never happen, but you never know!
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	err = os.Chown(RuntimeDir, os.Getuid(), os.Getgid())
	if err != nil {
		return nil, err
	}

	return open(UserRuntime(file), flag)
}

// open opens the given file with the appropriate flag and permission.
// The flag should be specified, depending on purpose. If O_CREATE is
// given, directories leading to the flag are also created.
//
//  O_RDONLY    open the file read-only.
//  O_WRONLY    open the file write-only.
//  O_RDWR      open the file read-write.
//  O_APPEND    append data to the file when writing.
//  O_CREATE    create a new file if none exists.
//  O_EXCL      used with O_CREATE, file must not exist
//  O_SYNC      open for synchronous I/O.
//  O_TRUNC     if possible, truncate file when opened.
func open(file string, flag int) (*os.File, error) {
	if file == "" {
		return nil, ErrInvalidPath
	}

	if flag&os.O_CREATE != 0 {
		// Check if we need to try to create a directory.
		err := MkdirAll(path.Dir(file))
		if err != nil {
			return nil, err
		}
	}

	return os.OpenFile(file, flag, 0700)
}

// MkdirAll creates dirpath if it does not already exist.
//
// Example:
//
//  xdg.MkdirAll(xdg.UserData("dromi"))
//  db, err := OpenDatabase(xdg.UserData("dromi/datbase.db"))
//
func MkdirAll(dirpath string) error {
	// TODO: am I swallowing err?
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		return os.MkdirAll(dirpath, os.ModeDir|0700)
	}
	return nil
}
