repoctl
========

**repoctl** is a program to help you manage your local pacman repository
(commonly on Arch Linux.) It is especially useful when used together with
[cower](https://github.com/falconindy/cower) (or similar programs) for
maintaining a local repository of AUR packages.

The history of this project is (for me) interesting; *repoctl* is the
successor to *repo-keep*, which is the successor to *repo-update*.
The first was written in Bash, then in C, and finally this one in Go.

The program *repoctl* is distributed under the MIT License (see `LICENSE`).

### Features
There are a few actions that can be specified on the command line, the most
important of which are:

  - `list` – List the packages that are physically present in the repository,
    optionally also in columnated format (like ls.)
  - `filter` – Filter the list of packages by certain criteria and print the
    names on their own lines. This can be combined well with other command
    line tools.
  - `status` – Show the current status of the repository, including pending
    packages, database inconsistencies, and AUR inconsistencies (such as,
    there is a newer version on AUR).
  - `update` – Update the database, removing old entries and files from the
    repository and adding new entries to the database.
  - `remove` – Remove files and entries from the database.

There are more features than this, have a look at the command line help for
more examples.

### Usage
Once you have created a configuration file (just run `repoctl` and it will
create one for you; you need to edit the config afterwards though), repo
is mostly used in the following way (see the Tips section for more):

    $ makepkg -c
      [...]
    $ repoctl update
      [...]


### Configuration File Example
The configuration file is located at `~/.repoctl.conf`, and is in the
[TOML](https://github.com/toml-lang/toml) format:

    # repo is the full path to the repository that will be managed by repoctl.
    # The packages that belong to the repository are assumed to lie in the
    # same folder.
    repo = "/srv/abs/atlas.db.tar.gz"

    # add_params is the set of parameters that will be passed to repo-add
    # when it is called. Specify one time for each parameter.
    #add_params = []

    # rm_params is the set of parameters that will be passed to repo-remove
    # when it is called. Specify one time for each parameter.
    #rm_params = []

### Tips
It makes most sense (to me) if you have a location where all your built
packages end up (see `/etc/makepkg.conf`). Then you would do something like
this:

    $ makepkg -c
      [...]
    $ repoctl update
      [...]

There is thus no copying of package files involved.

Here is something that I do a lot (except I have a script for the second
command):

    $ cower -dd $(repoctl filter outdated)
      [...]
    $ for dir in *; do cd $dir; makepkg -cs; cd ..; rm -rf $dir; done
      [...]
    $ repoctl update
      [...]

