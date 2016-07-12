repoctl
========

The **repoctl** program helps you manage your local repository of Pacman
packages (as found on Arch Linux and derivatives). It is especially useful when
used together with tools such as [cower](https://github.com/falconindy/cower),
that help you find and download AUR packages.

The repoctl program is distributed under the [MIT License](LICENSE).

### Installation

The recommended method is to install the `repoctl` package from [AUR](https://aur.archlinux.org/packages/repoctl),
as this package installs other useful files, such as the Zsh completion script.

Alternatively, if you have [Go](https://golang.org) installed:

```sh
go get -u github.com/cassava/repoctl
cd $GOPATH/src/github.com/cassava/repoctl
go install ./cmd/...
```

You may want to switch to the `devel` branch if you want the bleeding edge.

### Usage
Before you can use repoctl, you need to create a configuration file.
This tells repoctl where your local repository is, among other things.
Since no one really likes doing this step, repoctl can write a default
configuration for you. It will also tell you where it is writing the
configuration file, so you can change it at a later time.

Let's say you want your repository at `/srv/pkgs`,
and you want to name it `myrepo`. Then you would run:

```sh
repoctl new config /srv/pkgs/myrepo.db.tar.gz
```

Now, we can add and manipulate packages in the specified local repository.
You can see the currently active configuration by running `repoctl version`.
Note: repoctl will not create the directory if it does not already exists, so
make sure you do this at some point.

To add one or more packages to the repository, we can run:

```sh
repoctl add xbindkeys-1.8.6-1-x86_64.pkg.tar.xz rxvt-unicode-9.22-6-*.pkg.tar.xz
```

This command will add them to the directory and the database, and remove older
versions of the same package in the database.

You can also use repoctl to manage AUR packages. You can download one or more
AUR packages with:

```sh
cd ~/ibuildhere
repoctl down cantata-git rxvt-unicode-patched
```

These packages are then downloaded and extracted. Note: the `down` subcommand
currently does not fetch dependencies. If you have configured makepkg to put
these in your repository (see `PKGDEST` variable in `/etc/makepkg.conf`), then
you can update your repository database with:

```sh
repoctl update
```

You can check the status of your repository, including whether there are any
updates to your packages from AUR, with:

```sh
repoctl status -a
```

If you find you have a list of packages that have newer versions on AUR, you
can get them all in one go. If you are feeling adventurous, you can build
them in one go too:

```sh
cd ~/ibuildhere
repoctl down -u
for dir in *; do
  cd $dir
  makepkg -cs
  if [ $? -eq 0 ]; then
    repoctl add *.pkg.tar.xz
    cd ..
    rm -rf $dir
  else
    cd ..
  fi
done
```

You can pack that last bit on one line with:

```sh
for dir in *; do cd $dir; makepkg -cs && repoctl add *.pkg.tar.xz && cd ..  && rm -rf $dir || cd ..; done
```

If you set up `/etc/makepkg.conf` to put the built packages already in your
repository, then you can just run `repoctl update` instead of adding them
at each step.

These are not the only things that repoctl can do, to get a fuller picture, have
a look at the help, which you can always get on the command or any of the
subcommands with the `--help` flag or by running

```sh
repoctl help [cmd]
```

Enjoy!

### Tips

#### Packages on a remote filesystem
If you have a super fast internet connection and want your packages on a remote
server, you can get repoctl to play along with the `pre_action` and
`post_action` options in the configuration file:

```toml
# ...
pre_action  = "sshfs server:location ~/localmnt"
post_action = "fusermount -u ~/localmnt"
```

### Configuration File Example
The configuration file is normally located at `~/.config/repoctl/config.toml`,
and is in the [TOML](https://github.com/toml-lang/toml) format:

```toml
# repoctl configuration

# repo is the full path to the repository that will be managed by repoctl.
# The packages that belong to the repository are assumed to lie in the
# same folder.
repo = "/srv/abs/graphite.db.tar.gz"

# add_params is the set of parameters that will be passed to repo-add
# when it is called. Specify one time for each parameter.
add_params = [
  "-v"
]

# rm_params is the set of parameters that will be passed to repo-remove
# when it is called. Specify one time for each parameter.
rm_params = []

# ignore_aur is a set of package names that are ignored in conjunction
# with AUR related tasks, such as determining if there is an update or not.
ignore_aur = [
    "dropbox",
]

# backup specifies whether package files should be backed up or deleted.
# If it is set to false, then obsolete package files are deleted.
backup = false

# backup_dir specifies which directory backups are stored in.
# If a relative path is given, then it is relative to the repository.
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

# pre_action is a command that should be executed before doing anything
# with the repository, like reading or modifying it. Useful for mounting
# a remote filesystem.
pre_action = "sshfs host:www/pkgs.me ~/mnt"

# post_action is a command that should be executed before exiting.
post_action = "fusermount -u ~/mnt"
```

