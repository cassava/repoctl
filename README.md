Repo-Keep
=========

**repo-keep** is the successor to *repo-update*. repo-update was a supplement
to `repo-add`, which simplified updating local repositories, it was not
a supplement to `repo-remove` however.

This changes with *repo-keep*, which I shall also from here on refer to as
**repo**.

*repo-keep* is distributed under the new BSD License (see `LICENSE`).


### Features
Writing is hard, so to save time here is the (somewhat outdated) output of
`repo --help`:

    Usage: repo [OPTION...] <add|list|remove|update|sync> [PACKAGES ...]
    Manage local pacman repositories.

    Commands available:
      add <pkgname>    Add the package(s) with <pkgname> to the database by
                       finding in the same directory of the database the latest
                       file for that package (by file modification date),
                       deleting the others, and updating the database.
      list             List all the packages that are currently available.
      remove <pkgname> Remove the package with <pkgname> from the database, by
                       removing its entry from the database and deleting the files
                       that belong to it.
      update           Same as add, except scan and add changed packages.
      synchronize      Compare packages in the database to AUR for new versions.

    NOTE: In all of these cases, <pkgname> is the name of the package, without
    anything else. For example: pacman, and not pacman-3.5.3-1-i686.pkg.tar.xz

      -n, --noconfirm            Don't confirm file deletion
      -s, --soft                 Don't delete any files (n/a for: sync)
      -v, --verbose              Be loud and verbose
      -c, --config=CONFIG        Alternate configuration file
      -?, --help                 Give this help list
          --usage                Give a short usage message
      -V, --version              Print program version

    Mandatory or optional arguments to long options are also mandatory or optional
    for any corresponding short options.


### Usage
Once you have created a configuration file (just run `repo` and it will
create one for you; you need to edit the config afterwards though), repo
is mostly used in the following way (which is an example):

    $ makepkg -c
      [...]
    $ repo update


### Repo-Update Configuration File Example
The configuration file is located at `~/.repo.conf`.

    db_name = local.db.tar.gz
    db_dir = /home/abs/packages


### Limitations
Note that if you do the following, say with the program `aurget` (from
the AUR), the behavior may surprise you:

    $ aurget -Sb package1
      [...]
    $ aurget -Sb package2
      [...]
    $ repo add package1
    $ repo update

The last command will result in repo not finding any new packages,
because it compares the ages of files with the age of the database.


### Tips
repo-keep makes most sense (to me) if you have a location where all your
built packages end up (see `/etc/makepkg.conf`). Then you would do something
like this:

    $ makepkg -c
      [...]
    $ repo update
      [...]

