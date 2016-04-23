goulash/xdg
===========

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/goulash/xdg)
[![License: MIT](http://img.shields.io/badge/license-MIT-red.svg?style=flat-square)](http://opensource.org/licenses/MIT)

Package xdg provides an implementation of the XDG Base Directory Specification.

On initialization, the following package variables are set to their recommended
values, by reading the corresponding environment variables and falling back to
specification defaults if necessary.

    ConfigHome      // user configuration base directory, e.g. ~/.config
    DataHome        // user data files base directory, e.g. ~/.local/share
    CacheHome       // user cache files base directory, e.g. ~/.cache
    RuntimeDir      // user runtime files base directory, e.g. /run/user/1000
    ConfigDirs      // global configuration directories, e.g. /etc/xdg
    DataDirs        // global data files directories, e.g. /usr/local/share
    AllConfigDirs   // user and global configuration directories
    AllDataDirs     // user and global data directories

Initialization happens automatically, but can also be explicitely started with
the `Init` function. If no valid path can be constructed, the variable is left
blank or empty. If one of the required paths is blank or empty, the program
should fail. These variables should be treated as read-only; change them only
if you know what you are doing.

The package has four classes of functions, which should suffice for most needs:

    User*           // construct a valid path for user (config|data|...) files
    Find*           // find existing (config|data|...) files
    Merge*          // execute a function on each found (config|data) file
    Open*           // open or create a user (config|data|...) file

Only the `Open*` functions may alter the filesystem in any way: this is
restricted to creating XDG user base directories and files therein. Directories
in `ConfigDirs` and `DataDirs` are not modified.

The XDG Base Directory Specification, henceforth “the specification”, defines
several types of files: configuration, data, cache, and runtime files.
The specification can be found at:

    http://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html

## Configuration files

Configuration files are read from `ConfigHome` and from `ConfigDirs`;
they are only written in `ConfigHome`.

`ConfigHome` is a single base directory relative to which user-specific data
files should be written. This directory is defined by the environment
variable `$XDG_DATA_HOME`. If `$XDG_CONFIG_HOME` is not set, the default
`$HOME/.config` is used.

`ConfigDirs` is a set of preference ordered base directories relative to
which configuration files should be searched. This set of directories is
defined by the environment variable `$XDG_CONFIG_DIRS`. The directories in
`$XDG_CONFIG_DIRS` should be seperated with a colon `:`. If `$XDG_CONFIG_DIRS`
is not set, the default `/etc/xdg` is used.

`ConfigHomeDirs` combines `ConfigHome` and `ConfigDirs` into one preference
ordered set of directories.

## Data files

Data files are read from `DataHome` and from `DataDirs`;
they are only written in `DataHome`.

`DataHome` is a single base directory relative to which user-specific data files
should be written. This directory is defined by the environment variable
`$XDG_DATA_HOME`. If `$XDG_DATA_HOME` is not set, the default `$HOME/.local/share`
is used.

`DataDirs` is a set of preference ordered base directories relative to which data
files should be searched. This set of directories is defined by the environment
variable `$XDG_DATA_DIRS`. If `$XDG_CONFIG_DIRS` is not set, the default
`/usr/local/share:/usr/share` is used.

## Cache files

`CacheHome` is a single base directory relative to which user-specific
non-essential (cached) data should be written. This directory is defined by the
environment variable `$XDG_CACHE_HOME`.  If `$XDG_CACHE_HOME` is not set, the
default `$HOME/.cache` is used.

## Runtime files

`RuntimeDir` is a single base directory relative to which user-specific
runtime files and other file objects should be placed. This directory is
defined by the environment variable `$XDG_RUNTIME_DIR`. If `$XDG_RUNTIME_DIR`
is not set, the following method is used to find an appropriate directory:

    path.Join(os.TempDir(), fmt.Sprintf("xdg-%d", os.Getuid()))

This usually results in paths such as `/tmp/xdg-1000`. Normally, we expect
something along the lines of `/run/user/1000`.

In this implementation, we assume that the system takes care of removing the
XDG runtime directory at shutdown.

For more information, see the [documentation](http://godoc.org/github.com/goulash/xdg)! :-)
This package is licensed under the MIT license.
This package takes much inspiration from [adrg/xdg](https://github.com/adrg/xdg). Many Thanks.
