repoctl
========

**repoctl** is a program to help you manage your local pacman repository
(commonly on Arch Linux.) It is especially useful when used together with
[cower](https://github.com/falconindy/cower) (or similar programs) for
maintaining a local repository of AUR packages.

The program *repoctl* is distributed under the MIT License (see `LICENSE`).

### Features
There are a few actions that can be specified on the command line, the most
important of which are:

  - `list` – List the packages that are physically present in the repository,
    optionally also in columnated format (like ls.)
  - `status` – Show the current status of the repository, including pending
    packages, database inconsistencies, and AUR inconsistencies (such as,
    there is a newer version on AUR).
  - `down` – Download out-of-date packages from AUR that are in the repository.
  - `add` – Add packages to the database.
  - `update` – Update the database, removing old entries and files from the
    repository and adding new entries to the database.
  - `remove` – Remove files and entries from the database.

There are more features than this, have a look at the command line help for
more examples.

### Installation
Either install `repoctl` from AUR using your preferred method, or if you have
`go` already installed, then:

```
go get github.com/cassava/repoctl
```

### Usage
Once you have created a configuration file (just run `repoctl new` and it will
create one for you; you may need to edit the config afterwards though). Once you
have a repository, you could use repoctl in the following way
(see the Tips section for more):

Download all the packages that need to be updated.

    $ repoctl down -u
      [...]
    $ for dir in *; do cd $dir; makepkg -cs && repoctl add *.pkg.tar.xz && cd ..
    && rm -rf $dir || cd ..; done

If you set up `/etc/makepkg.conf` to put the built packages already in your
repository, then you can just run `repoctl update`.

### Configuration File Example
The configuration file is located at `~/.config/repoctl/config.toml`, and is in the
[TOML](https://github.com/toml-lang/toml) format:

```
# repoctl configuration

# repo is the full path to the repository that will be managed by repoctl.
# The packages that belong to the repository are assumed to lie in the
# same folder.
repo = "/srv/abs/graphite.db.tar.gz"

# add_params is the set of parameters that will be passed to repo-add
# when it is called. Specify one time for each parameter.
#add_params = []

# rm_params is the set of parameters that will be passed to repo-remove
# when it is called. Specify one time for each parameter.
#rm_params = []

# ignore_aur is a set of package names that are ignored in conjunction
# with AUR related tasks, such as determining if there is an update or not.
ignore_aur = [
    "colemak-bm",
]

# backup specifies whether package files should be backed up or deleted.
# If it is set to false, then obsolete package files are deleted.
backup = false

# backup_dir specifies which directory backups are stored in.
# If a relative path is given, then 
backup_dir = "backup/"

# interactive specifies that repoctl should ask before doing anything
# destructive.
interactive = false

# columnate specifies that listings should be in columns rather than
# in lines. This only applies to the list command.
columnate = true

# quiet specifies whether repoctl should print more information or less.
# I prefer to know what happens, but if you don't like it, you can change it.
quiet = false
```

### Tips
It makes most sense (to me) if you have a location where all your built
packages end up (see `/etc/makepkg.conf`). Then you would do something like
this:

    $ makepkg -c
      [...]
    $ repoctl update
      [...]
