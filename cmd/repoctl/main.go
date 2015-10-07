// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/cassava/repoctl"
	"github.com/spf13/cobra"
)

// Reset -------------------------------------------------------------

var ResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "recreate repository database",
	Long: `Delete the repository database and re-add all packages in repository.
    
  Essentially, this command deletes the repository database and
  recreates it by running the update command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := Repo.Reset(nil)
		dieOnError(err)
	},
}

// Add ---------------------------------------------------------------

var movePackages bool

func init() {
	AddCmd.Flags().BoolVarP(&movePackages, "move", "m", false, "move packages into repository")
}

var AddCmd = &cobra.Command{
	Use:   "add <pkgfile...>",
	Short: "copy and add packages to the repository",
	Long: `Add (and copy if necessary) the package file to the repository.

  Similarly to the repo-add script, this command copies the package
  file to the repository (if not already there) and adds the package to
  the database. Exactly this package is added to the database, this
  allows you to downgrade a package in the repository.

  Any other package files in the repository are deleted or backed up,
  depending on whether the backup option is given. If the backup option
  is given, the "obsolete" package files are moved to a backup
  directory of choice.

  Note: since version 0.14, the semantic meaning of this command has
        changed. See the update command for the old behavior.
`,
	Example: `  repoctl add -m ./fairsplit-1.0.pkg.tar.gz`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if movePackages {
			err = Repo.Move(nil, args...)
		} else {
			err = Repo.Copy(nil, args...)
		}
		dieOnError(err)
	},
}

// Update ------------------------------------------------------------

var UpdateCmd = &cobra.Command{
	Use:   "update [pkgname...]",
	Short: "update database in repository",
	Long: `Update database by adding packages and dispatching obsolete files.

  Package files that are newer than the registered versions in the
  database are added to the database; entries in the database which
  have no files are removed.

  If no package names are given, the entire repository is scanned for
  updates.
`,
	Example: `  repoctl update fairsplit`,
	Run: func(cmd *cobra.Command, args []string) {
		err := Repo.Update(nil, args...)
		dieOnError(err)
	},
}

// Remove ------------------------------------------------------------

var RemoveCmd = &cobra.Command{
	Use:     "remove <pkgname...>",
	Aliases: []string{"rm"},
	Short:   "remove and delete packages from the database",
	Long: `Remove and delete the package files from the repository.

  This command removes the specified package from the repository
  database, and deletes any associated package files, unless the backup
  option is given, in which case the package files are moved to the
  backup directory.
`,
	Example: `  repoctl rm fairsplit`,
	Run: func(cmd *cobra.Command, args []string) {
		err := Repo.Remove(nil, args...)
		dieOnError(err)
	},
}

// Down --------------------------------------------------------------

var (
	downDest     string
	downClobber  bool
	downExtract  bool
	downUpgrades bool
	downAll      bool
)

func init() {
	DownCmd.Flags().StringVarP(&downDest, "dest", "d", "", "output directory for tarballs")
	DownCmd.Flags().BoolVarP(&downClobber, "clobber", "b", false, "delete conflicting files and folders")
	DownCmd.Flags().BoolVarP(&downExtract, "extract", "e", true, "extract the downloaded tarballs")
	DownCmd.Flags().BoolVarP(&downUpgrades, "upgrades", "u", false, "download tarballs for all upgrades")
	DownCmd.Flags().BoolVarP(&downAll, "all", "a", false, "download tarballs for all packages in database")
}

var DownCmd = &cobra.Command{
	Use:     "down [pkgname...]",
	Aliases: []string{"download"},
	Short:   "download and extract tarballs from AUR",
	Long: `Download and extract tarballs from AUR for given packages.
Alternatively, all packages, or those with updates can be downloaded.
Options specified are additive, not exclusive.

By default, tarballs are deleted after being extracted, and are placed
in the current directory.
`,
	Example: `  repoctl down -u`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if downAll {
			names, err := Repo.ReadAUR(nil)
			dieOnError(err)
			err = Repo.Download(nil, downDest, downExtract, downClobber, names...)
		} else if downUpgrades {
			err = Repo.DownloadUpgrades(nil, downDest, downExtract, downClobber, args...)
		} else {
			err = Repo.Download(nil, downDest, downExtract, downClobber, args...)
		}
		dieOnError(err)
	},
}

// Version -----------------------------------------------------------

type programInfo struct {
	Name      string
	Author    string
	Email     string
	Version   string
	Date      string
	Homepage  string
	Copyright string
	License   string
}

const versionTmpl = `{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at
    {{.Homepage}}
Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.
`

var progInfo = programInfo{
	Name:      "repoctl",
	Author:    "Ben Morgan",
	Email:     "neembi@gmail.com",
	Version:   "0.14",
	Date:      "6 October 2015",
	Copyright: "2015",
	Homepage:  "https://github.com/cassava/repoctl",
	License:   "MIT",
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version and date information",
	Long:  "Show the official version number of repoctl, as well as the release date.",
	func(cmd *cobra.Command, args []string) {
		template.Must(template.New("version").Parse(versionTmpl)).Execute(os.Stdout, progInfo)
	},
}

// Main --------------------------------------------------------------

var Repo *repoctl.Repo

var repoctlCmd = &cobra.Command{
	Use:   "repoctl",
	Short: "manage local Pacman repositories",
	Long: `Repoctl helps manage local Pacman repositories, and acts in particular as
a supplement to the repo-add and repo-remove tools that come with Pacman.

Whether compiling and installing from AUR every time is not what you want,
or if you want to host your own repository, repoctl is right for you.

Note that in all of these commands, the following terminology is used:

    pkgname: is the name of the package, e.g. pacman
    pkgfile: is the path to a package file, e.g. pacman-3.5.3-i686.pkg.tar.xz
`,
}

func addConfFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&Conf.Backup, "backup", "b", Conf.Backup, "backup obsolete files instead of deleting")
	cmd.PersistentFlags().StringVarP(&Conf.BackupDir, "backup-dir", "B", Conf.BackupDir, "backup directory relative to repository path")
	cmd.PersistentFlags().BoolVarP(&Conf.Columnate, "columns", "s", Conf.Columnate, "show items in columns rather than lines")
	cmd.PersistentFlags().BoolVarP(&Conf.Interactive, "interactive", "i", Conf.Interactive, "ask before doing anything destructive")
	cmd.PersistentFlags().BoolVarP(&Conf.Quiet, "quiet", "q", Conf.Quiet, "show minimal amount of information")
	cmd.PersistentFlags().BoolVar(&Conf.Debug, "debug", Conf.Debug, "show unnecessary debugging information")
}

func addCommands(cmd *cobra.Command) {
	cmd.AddCommand(StatusCmd)
	cmd.AddCommand(ListCmd)
	cmd.AddCommand(FilterCmd)
	cmd.AddCommand(NewCmd)
	cmd.AddCommand(AddCmd)
	cmd.AddCommand(RemoveCmd)
	cmd.AddCommand(UpdateCmd)
	cmd.AddCommand(ResetCmd)
	cmd.AddCommand(DownCmd)
	cmd.AddCommand(VersionCmd)
}

// main loads the configuration and executes the primary command.
func main() {
	err := Conf.MergeAll()
	if err != nil {
		// We didn't manage to load any configuration, which means that repoctl
		// is unconfigured. There are some commands that work nonetheless, so
		// we can't stop now -- which is why we don't os.Exit(1).
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
	}

	// Arguments from the command line override the configuration file,
	// so we have to add the flags after loading the configuration.
	//
	// TODO: Maybe in the future we will make it possible to specify the
	// configuration file via the command line; right now it is not a priority.
	addConfFlags(repoctlCmd)
	addCommands(repoctlCmd)

	// All the commands rely on Repo being set correctly. Do that now.
	Repo = repoctl.New(Conf.Repository)
	Repo.Backup = Conf.Backup
	Repo.BackupDir = Conf.BackupDir
	Repo.AddParameters = Conf.AddParameters
	Repo.RemoveParameters = Conf.RemoveParameters
	Repo.Error = os.Stderr
	if Conf.Quiet {
		Repo.Info = nil
	}
	if Conf.Debug {
		Repo.Info = os.Stdout
		Repo.Debug = os.Stdout
	}

	repoctlCmd.Execute()
}
