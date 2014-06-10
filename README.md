repoctl
========

**repoctl** is a program to help you manage your local Arch Linux
repository. It is especially useful when used together with
[cower](https://github.com/falconindy/cower) (or similar programs) for
maintaining a local repository of AUR packages.

The history is interesting; *repoctl* is the successor to *repo-keep*,
which is the successor to *repo-update*. The first was written in Bash,
then in C, and finally in Go.

The program *repoctl* is distributed under the MIT License (see `LICENSE`).

### Features
There are a few actions that can be specified on the command line, the most
important of which are:

 - `list` – List the packages that are physically present in the repository,
   optionally also in columnated format (like ls.)
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
is mostly used in the following way (which is an example):

    $ makepkg -c
      [...]
    $ repoctl update


### Configuration File Example
The configuration file is located at `~/.repo.conf`.

    [atlas]
    repository = /srv/abs/atlas.db.tar.gz

### Tips
It makes most sense (to me) if you have a location where all your built
packages end up (see `/etc/makepkg.conf`). Then you would do something like
this:

    $ makepkg -c
      [...]
    $ repoctl update
      [...]

