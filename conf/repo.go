// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package conf

import (
	"os"

	"github.com/cassava/repoctl"
)

// Return a Repo object set according to the configuration.
func (c *Configuration) Repo() *repoctl.Repo {
	if c.Unconfigured {
		return nil
	}

	r := repoctl.New(c.Repository)
	r.Backup = c.Backup
	r.BackupDir = c.BackupDir
	r.IgnoreAUR = c.IgnoreAUR
	r.AddParameters = c.AddParameters
	r.RemoveParameters = c.RemoveParameters
	r.Error = os.Stderr
	if c.Quiet {
		r.Info = nil
	}
	if c.Debug {
		r.Info = os.Stdout
		r.Debug = os.Stdout
	}

	return r
}
