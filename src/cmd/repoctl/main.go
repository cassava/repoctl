// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
	flag "gopkg.in/ogier/pflag.v0"
)

var RepoctlCmd = &cobra.Command{
	Use:   "repoctl",
	Short: "manage local Pacman repositories",
	Long: `repoctl helps manage local Pacman repositories, and acts in particular as
a supplement to the repo-add and repo-remove tools that come with Pacman.
Whether compiling and installing from AUR every time is not what you want,
or if you want to host your own repository, repoctl is right for you.

Note that in all of these commands, the following terminology is used:

    pkgname: is the name of the package, e.g. pacman
    pkgfile: is the path to a package file, e.g. pacman-3.5.3-i686.pkg.tar.xz
`,
	Run: repoctl,
}

var (
	Repository       string
	AddParameters    []string
	RemoveParameters []string
	IgnoreAUR        map[string]bool
	BackupDir        string
	Backup           bool
	Interactive      bool
	Quiet            bool

	// When Debug is specified, it presides over Quiet.
	// This allows it to override a possible default value of Quiet.
	Debug bool
)

func init() {
	IgnoreAUR = make(map[string]bool)

	RepoctlCmd.PersistentFlags().StringVarP(&BackupDir, "backup-dir", "d", "backup", "backup directory relative to repository path")
	RepoctlCmd.PersistentFlags().BoolVarP(&Backup, "backup", "b", false, "backup obsolete files instead of deleting")
	RepoctlCmd.PersistentFlags().BoolVarP(&Interactive, "interactive", "i", false, "ask before doing anything destructive")
	RepoctlCmd.PersistentFlags().BoolVarP(&Columnate, "columns", "s", false, "show items in columns rather than lines")
	RepoctlCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "show minimal amount of information")
	RepoctlCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "show unnecessary debugging information")
}

func init() {
	RepoctlCmd.AddCommand(StatusCmd) // TODO
	RepoctlCmd.AddCommand(ListCmd)   // TODO
	RepoctlCmd.AddCommand(FilterCmd) // TODO
	RepoctlCmd.AddCommand(NewCmd)    // TODO
	RepoctlCmd.AddCommand(AddCmd)    // TODO
	RepoctlCmd.AddCommand(RemoveCmd) // TODO
	RepoctlCmd.AddCommand(UpdateCmd) // TODO
	RepoctlCmd.AddCommand(ResetCmd)  // TODO
	RepoctlCmd.AddCommand(DownCmd)   // TODO
	RepoctlCmd.AddCommand(VersionCmd)
	RepoctlCmd.AddCommand(HelpCmd) // TODO
}

// TODO: At this stage, the most important thing is to investigate how to use Viper in this project.
func repoctl(cmd *cobra.Command, args []string) {
	panic("not implemented")
}

var defaultConfigPath = path.Join(os.Getenv("HOME"), ".repoctl.conf")

// Config contains all the configuration flags, variables, and arguments that
// are needed for the various actions.
type Config struct {
	// ConfigFile stores the name of the configuration file from which this
	// configuration was loaded from, otherwise it is empty.
	ConfigFile string

	// Repository is the absolute path to the database. We assume that this is
	// also where the packages are. The variables database and path are constructed
	// from this.
	Repository string
	database   string
	path       string
	// AddParameters are parameters to add to the repo-add command line.
	AddParameters []string
	// RemoveParameters are parameters to add to the repo-remove command line.
	RemoveParameters []string
	// Packages to ignore when doing AUR related tasks.
	IgnoreAUR map[string]bool

	// Quiet causes less information to be printed than usual.
	Quiet bool
	// Columnate causes items to be printed in columns rather than lines.
	Columnate bool

	// Versioned causes packages to be printed with version information.
	Versioned bool
	// Mode can be either "count", "filter", or "mark" (which is the default
	// if no match is found.
	Mode string
	// Pending marks packages that need to be added to the database,
	// as well as packages that are in the database but are not available.
	Pending bool
	// Duplicates marks the number of obsolete packages for each package.
	Duplicates bool
	// Installed marks whether packages are locally installed or not.
	Installed bool
	// Synchronize marks which packages have newer versions on AUR.
	Synchronize bool

	// Interactive requires confirmation before deleting and changing the
	// repository database.
	Interactive bool
	// Backup causes older packages to be backed up rather than deleted.
	// For this, the files are given the suffix ".bak".
	Backup bool

	// Arguments contains the arguments given on the commandline.
	Args []string
}

func NewConfig() *Config {
	return &Config{
		IgnoreAUR: make(map[string]bool),
	}
}

// Usage prints the help message for the program.
func Usage(*Config) error {
	fmt.Printf("%s %s (%s)", progName, progVersion, progDate)
	fmt.Print(`
Manage local pacman repositories.

Commands available:

  list             List packages that belong to the managed repository.
  ls               Options available are:
                    -v --versions   show package versions along with name
                    -d --duplicates mark packages with duplicate package files
                    -p --pending    mark pending changes to the database
                    -l --installed  mark packages that are locally installed
                    -o --outdated   mark packages that are newer in AUR
                    -a --all        same as -vpdlu

  filter <crit...> Filter list of packages by one or more criteria;
                   run without any criteria for help.

  status           Show pending changes to the database and packages that can
                   be updated.

  add <pkgname>    Add the latest package(s) with <pkgname> to the database
                   and delete all obsolete package files.

  remove <pkgname> Remove the package(s) with <pkgname> from the database and
  rm               delete all the corresponding package files.

  update           Automatically scan the repository for changes and update
                   by changing the database and deleting obsolete package files.

  reset            Reset the repository database by removing it and adding all
                   up-to-date packages while deleting obsolete package files.

                   Options available to add, remove, update, and reset are:
                    -i --interactive  ask before doing anything destructive
                    -b --backup       backup obsolete package files instead of
                                      deleting; packages are put into backup/

  down <pkgname>   Download and extract tarballs from AUR for given packages.
                   Alternatively, all packages, or those with updates can be
                   downloaded. The options below are additive, not exclusive.
                    -u --updates    download tarballs for updates
                    -a --all        download tarballs for all packages
                    -n --no-extract do not extract the tarballs
                    -t --rewrite    delete conflicting folders

  help             Show the usage for repoctl. Synonym for
  usage             repoctl --help

NOTE: In all of these cases, <pkgname> is the name of the package, without
anything else. For example: pacman, and not pacman-3.5.3-1-i686.pkg.tar.xz
Multiple packages are usually accepted, separated by spaces.

General options available are:

 -h --help      show this usage message
 -q --quiet     only show information when absolutely necessary
 -s --columns   show items in columns rather than lines
 -c --config    configuration file to load settings from
    --repo      path to repository and database, such as
                "/srv/abs/atlas.db.tar.gz"
`)

	return nil
}

// ReadConfig reads a configuration from the command line arguments.
func ReadConfig() (conf *Config, cmd Action, err error) {
	var allListOptions bool
	var showHelp bool
	conf = NewConfig()

	flag.StringVarP(&conf.ConfigFile, "config", "c", defaultConfigPath, "configuration file to load settings from")
	flag.StringVar(&conf.Repository, "repo", "", "path to repository and database")

	// TODO: Implement --ignore=pkg1,pkg2,...

	flag.BoolVarP(&showHelp, "help", "h", false, "show this usage message")

	// List options
	flag.BoolVarP(&conf.Versioned, "versioned", "v", false, "show package versions along with name")
	flag.BoolVarP(&conf.Pending, "pending", "p", false, "mark pending changes to the database")
	flag.BoolVarP(&conf.Duplicates, "duplicates", "d", false, "mark packages with duplicate package files")
	flag.BoolVarP(&conf.Installed, "installed", "l", false, "mark packages that are locally installed")
	flag.BoolVarP(&conf.Synchronize, "outdated", "o", false, "mark packages that are newer in AUR")
	flag.BoolVarP(&allListOptions, "all", "a", false, "all information; same as -vpdlo")

	flag.BoolVarP(&conf.Interactive, "interactive", "i", false, "ask before doing anything destructive")
	flag.BoolVarP(&conf.Backup, "backup", "b", false, "backup obsolete package files instead of deleting")

	flag.Usage = func() { Usage(nil) }
	flag.Parse()

	if showHelp {
		return nil, Usage, nil
	} else if len(flag.Args()) == 0 {
		return nil, Usage, errors.New("no action specified on command line")
	}

	// Read config file.
	var isDefault bool
	if ex, _ := osutil.FileExists(conf.ConfigFile); ex {
		rc, err := ReadRepoConfig(conf.ConfigFile)
		if err != nil {
			return nil, nil, err
		}
		isDefault = rc.Default
		rc.MergeIntoConfig(conf)
	} else {
		fmt.Fprintf(os.Stderr, "Warning: creating missing config file %q.\n", conf.ConfigFile)
		rp := "/srv/abs/atlas.db.tar.gz"
		if conf.Repository != "" {
			rp = conf.Repository
		}
		RepoConfig{Repo: rp}.WriteDefault(conf.ConfigFile)
	}

	// Fail if we still don't have repository information, or if config file has not been updated.
	if isDefault {
		return nil, nil, fmt.Errorf("please edit configuration file %q before running repoctl", conf.ConfigFile)
	}
	if conf.Repository == "" {
		return nil, nil, fmt.Errorf("missing repository information; set in %q!", conf.ConfigFile)
	}

	conf.path = path.Dir(conf.Repository)
	conf.database = path.Base(conf.Repository)
	if allListOptions {
		conf.Versioned = true
		conf.Pending = true
		conf.Duplicates = true
		conf.Installed = true
		conf.Synchronize = true
	}
	conf.Args = flag.Args()[1:]
	cmd, ok := actions[flag.Arg(0)]
	if !ok {
		return nil, Usage, errors.New("unrecognized action " + flag.Arg(0))
	}

	return conf, cmd, nil
}

func main() {
	conf, cmd, err := ReadConfig()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println()
		Usage(nil)
		os.Exit(1)
	}

	cmd(conf)
}
