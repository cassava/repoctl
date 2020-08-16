// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package conf

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cassava/repoctl/pacman/alpm"
)

// Profile doubles as configuration file format and the global configuration set.
type Profile struct {
	// Repository is the absolute path to the database. We assume that this is
	// also where the packages are. The variables database and path are constructed
	// from this.
	Repository string `toml:"repo"`
	database   string
	repodir    string

	// AddParameters are parameters to add to the repo-add command line.
	AddParameters []string `toml:"add_params"`
	// RemoveParameters are parameters to add to the repo-remove command line.
	RemoveParameters []string `toml:"rm_params"`
	// Packages to ignore when doing AUR related tasks.
	IgnoreAUR []string `toml:"ignore_aur"`
	// Require signatures for packages that are added to the database.
	RequireSignature bool `toml:"require_signature"`

	// Backup causes older packages to be backed up rather than deleted.
	Backup bool `toml:"backup"`
	// BackupDir specifies where old packages are backed up to.
	BackupDir string `toml:"backup_dir"`
	// Interactive requires confirmation before deleting and changing the
	// repository database.
	Interactive bool `toml:"interactive"`

	// PreAction and PostAction are run every time that the database or
	// filesystem is accessed.
	PreAction  string `toml:"pre_action"`
	PostAction string `toml:"post_action"`
}

func DefaultProfile() *Profile {
	return &Profile{
		BackupDir: "backup/",
	}
}

// NewProfile creates a new default configuration with repo as
// the repository database.
//
// If repo is invalid (because it is absolute), nil is returned.
// We don't check for existance, because at this point, it might
// not exist yet.
func NewProfile(repo string) *Profile {
	if !path.IsAbs(repo) {
		return nil
	}
	return &Profile{
		Repository: repo,
		database:   path.Base(repo),
		repodir:    path.Dir(repo),

		AddParameters:    make([]string, 0),
		RemoveParameters: make([]string, 0),
		IgnoreAUR:        make([]string, 0),
	}
}

// Init should be called every time after changing c.Repository.
// This ensures that the configuration is both valid and consistent.
//
// If the repository path is not absolute, ErrRepoNotAbs is returned.
func (p *Profile) Init() error {
	if !path.IsAbs(p.Repository) {
		return ErrRepoNotAbs
	}
	p.database = path.Base(p.Repository)
	p.repodir = path.Dir(p.Repository)

	// Perform a check on database extension.
	if !alpm.HasDatabaseFormat(p.database) {
		fmt.Fprintf(os.Stderr, "Warning: Specified repository database %q has an unexpected extension.\n", p.database)
		fmt.Fprintf(os.Stderr, "         It should conform to this pattern: .db.tar.(zst|xz|gz|bz2).\n")
		base := filepath.Base(p.database)
		if i := strings.IndexRune(base, '.'); i != -1 {
			base = base[:i]
		}
		fmt.Fprintf(os.Stderr, "         For example: %s.db.tar.zst\n", filepath.Join(filepath.Dir(p.database), base))
	}

	return nil
}
