// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"text/template"

	"github.com/cassava/repoctl"
	"github.com/cassava/repoctl/conf"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/aur"
	"github.com/goulash/pacman/meta"
	"github.com/goulash/pacman/pkgutil"
	"github.com/goulash/pr"
	"github.com/spf13/cobra"
	"github.com/spf13/nitro"
)

// Repo lets us use the repoctl library to do the most of the work.
var Repo *repoctl.Repo

// Conf loads and stores the configuration (apart from command line
// configuration) of this program, including where the repository is.
var Conf = conf.Default()

// Timer
var Timer *nitro.B

// Colorizer is a pr.Colorizer instance, which you are able to customize.
// In particular, it may be useful to turn off colorization when the
// terminal does not support it, or when redirecting output:
//
//	if colorState == "auto" {
//		dungeon.Colorizer.SetFile(os.Stdout)
//	} else if colorState == "always" {
//		dungeon.Colorizer.SetEnabled(true)
//	} else if colorState == "never" {
//		dungeon.Colorizer.SetEnabled(false)
//	}
var col = pr.NewColorizer()

// Status ------------------------------------------------------------

var (
	statusAUR     bool
	statusMissing bool
)

func init() {
	StatusCmd.Flags().BoolVarP(&statusAUR, "aur", "a", false, "check AUR for upgrades")
	StatusCmd.Flags().BoolVarP(&statusMissing, "missing", "m", false, "highlight packages missing in AUR")
}

var StatusCmd = &cobra.Command{
	Use:   "status [--aur]",
	Short: "show pending changes and packages that can be upgraded",
	Long: `Show pending changes to the database and packages that can be updated.

  In particular, the following is shown:

    - obsolete package files that can be deleted (or backed up)
    - database entries that should be deleted (no package files)
    - database entries that can be updated/added (new package files)
    - packages unavailable in AUR
    - packages with updates in AUR
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmd.Usage()
			os.Exit(1)
		}

		dieOnError(Init())
		col.Printf("On repo @{!y}%s\n\n", Repo.Name())

		Timer.Step("read repository")
		pkgs, err := Repo.ReadMeta(nil)
		dieOnError(err)
		if statusAUR {
			err = pkgs.ReadAUR()
			if err != nil && !aur.IsNotFound(err) {
				dieOnError(err)
			}
		}

		// We assume that there is nothing to do, and if there is,
		// then this is set to false.
		var nothing = true

		Timer.Step("output")
		for _, p := range pkgs {
			var flags []string
			if p.HasUpgrade() {
				flags = append(flags, col.Sprintf("@gupgrade(@|%s->%s@g)", p.Version(), p.AUR.Version))
			}
			if p.HasUpdate() {
				flags = append(flags, col.Sprintf("@gupdate(@|%s->%s@g)", p.VersionRegistered(), p.Version()))
			}
			if !p.HasFiles() {
				flags = append(flags, col.Sprint("@rremoval"))
			}
			if o := p.Obsolete(); len(o) > 0 {
				flags = append(flags, col.Sprintf("@yobsolete(@|%d@y)", len(o)))
			}
			if statusMissing && p.AUR == nil {
				flags = append(flags, col.Sprint("@y!aur"))
			}

			if len(flags) > 0 {
				nothing = false
				fmt.Printf("\t%s:", p.Name)
				for _, f := range flags {
					fmt.Printf(" %s", f)
				}
				fmt.Println()
			}
		}

		if nothing {
			fmt.Println("Everything up-to-date.")
		}
	},
}

// List --------------------------------------------------------------

var (
	// Versioned causes packages to be printed with version information.
	listVersioned bool
	// Mode can be either "count", "filter", or "mark" (which is the default
	// if no match is found.
	listMode string
	// Pending marks packages that need to be added to the database,
	// as well as packages that are in the database but are not available.
	listPending bool
	// Duplicates marks the number of obsolete packages for each package.
	listDuplicates bool
	// Installed marks whether packages are locally installed or not.
	listInstalled bool
	// Synchronize marks which packages have newer versions on AUR.
	listSynchronize bool
	// Same as all of the above.
	listAllOptions bool
)

func init() {
	ListCmd.Flags().BoolVarP(&listVersioned, "versioned", "v", false, "show package versions along with name")
	ListCmd.Flags().BoolVarP(&listPending, "pending", "p", false, "mark pending changes to the database")
	ListCmd.Flags().BoolVarP(&listDuplicates, "duplicates", "d", false, "mark packages with duplicate package files")
	ListCmd.Flags().BoolVarP(&listInstalled, "installed", "l", false, "mark packages that are locally installed")
	ListCmd.Flags().BoolVarP(&listSynchronize, "outdated", "o", false, "mark packages that are newer in AUR")
	ListCmd.Flags().BoolVarP(&listAllOptions, "all", "a", false, "all information; same as -vpdlo")
}

var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list packages that belong to the managed repository",
	Long: `List packages that belong to the managed repository.

  Note that they don't need to be registered with the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		dieOnError(Init())
		if len(args) > 0 {
			cmd.Usage()
			os.Exit(1)
		}

		if listAllOptions {
			listVersioned = true
			listPending = true
			listDuplicates = true
			listInstalled = true
			listSynchronize = true
		}

		pkgs, err := Repo.ListMeta(nil, listSynchronize, func(mp pacman.AnyPackage) string {
			p := mp.(*meta.Package)
			if listPending && !p.HasFiles() {
				return fmt.Sprintf("-%s-", p.Name)
			}

			buf := bytes.NewBufferString(p.Name)
			if listPending && p.HasUpdate() {
				buf.WriteRune('*')
			}
			if listVersioned {
				buf.WriteRune(' ')
				buf.WriteString(p.Version())
			}
			if listSynchronize {
				ap := p.AUR
				if ap == nil {
					buf.WriteString(" <?>")
				} else if pacman.PkgNewer(ap, p) {
					if listVersioned {
						buf.WriteString(" -> ")
						buf.WriteString(ap.Version)
					} else {
						buf.WriteString(" <!>")
					}
				} else if pacman.PkgOlder(ap, p) {
					if listVersioned {
						buf.WriteString(" <- ")
						buf.WriteString(ap.Version)
					} else {
						buf.WriteString(" <*>")
					}
				}
			}
			if listDuplicates && len(p.Files)-1 > 0 {
				buf.WriteString(fmt.Sprintf(" (%v)", len(p.Files)-1))
			}

			return buf.String()
		})
		dieOnError(err)

		// Print packages to stdout
		printSet(pkgs, "", Conf.Columnate)
	},
}

// Reset -------------------------------------------------------------

var ResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "recreate repository database",
	Long: `Delete the repository database and re-add all packages in repository.

  Essentially, this command deletes the repository database and
  recreates it by running the update command.
`,
	Run: func(cmd *cobra.Command, args []string) {
		dieOnError(Init())
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
		dieOnError(Init())
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
		dieOnError(Init())
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
		dieOnError(Init())
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
	DownCmd.Flags().BoolVarP(&downClobber, "clobber", "l", false, "delete conflicting files and folders")
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
		dieOnError(Init())
		var err error
		if downAll {
			names, err := Repo.ReadNames(nil)
			dieOnError(err)
			err = Repo.Download(nil, downDest, downExtract, downClobber, pkgutil.Map(names, pkgutil.PkgName)...)
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
	Run: func(cmd *cobra.Command, args []string) {
		template.Must(template.New("version").Parse(versionTmpl)).Execute(os.Stdout, progInfo)
	},
}

// Main --------------------------------------------------------------

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
	cmd.PersistentFlags().BoolVar(&nitro.AnalysisOn, "step-analysis", false, "display memory and timing of different steps")
	cmd.PersistentFlags().BoolVarP(&Conf.Backup, "backup", "b", Conf.Backup, "backup obsolete files instead of deleting")
	cmd.PersistentFlags().StringVarP(&Conf.BackupDir, "backup-dir", "B", Conf.BackupDir, "backup directory relative to repository path")
	cmd.PersistentFlags().BoolVarP(&Conf.Columnate, "columns", "s", Conf.Columnate, "show items in columns rather than lines")
	cmd.PersistentFlags().BoolVarP(&Conf.Quiet, "quiet", "q", Conf.Quiet, "show minimal amount of information")
	cmd.PersistentFlags().BoolVar(&Conf.Debug, "debug", Conf.Debug, "show unnecessary debugging information")
	cmd.PersistentFlags().StringVar(&Conf.Color, "color", Conf.Color, "when to use color (auto|never|always)")
}

func addCommands(cmd *cobra.Command) {
	cmd.AddCommand(StatusCmd)
	cmd.AddCommand(ListCmd)
	//	cmd.AddCommand(FilterCmd)
	cmd.AddCommand(NewCmd)
	cmd.AddCommand(AddCmd)
	cmd.AddCommand(RemoveCmd)
	cmd.AddCommand(UpdateCmd)
	cmd.AddCommand(ResetCmd)
	cmd.AddCommand(DownCmd)
	cmd.AddCommand(VersionCmd)
}

// Init makes sure that Conf is configured, and sets Repo up.
func Init() error {
	if Conf.Color == "auto" {
		col.SetFile(os.Stdout)
	} else if Conf.Color == "always" {
		col.SetEnabled(true)
	} else if Conf.Color == "never" {
		col.SetEnabled(false)
	}

	if Conf.Unconfigured {
		return errors.New("repoctl is unconfigured, please create configuration")
	}
	Repo = Conf.Repo()
	return nil
}

// main loads the configuration and executes the primary command.
func main() {
	Timer = nitro.Initialize()

	Timer.Step("read configuration")
	err := Conf.MergeAll()
	if err != nil {
		// We didn't manage to load any configuration, which means that repoctl
		// is unconfigured. There are some commands that work nonetheless, so
		// we can't stop now -- which is why we don't os.Exit(1).
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
	}

	Timer.Step("initialize main")
	// Arguments from the command line override the configuration file,
	// so we have to add the flags after loading the configuration.
	//
	// TODO: Maybe in the future we will make it possible to specify the
	// configuration file via the command line; right now it is not a priority.
	addConfFlags(repoctlCmd)
	addCommands(repoctlCmd)

	repoctlCmd.Execute()
}
