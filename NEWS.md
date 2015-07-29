Repoctl Releases
================

## Version 0.14 (-)
This release changes the AUR lookup functionality to use AUR4 instead of AUR.
When https://aur4.archlinux.org is the same as https://aur.archlinux.org, we
will revert this change. Additionally, we add the command `down` for
downloading PKGBUILD tarballs.

  - New: repoctl learned command `down`
  - Update: AUR retrieval has been improved
  - Bugfix: license information correct (was BSD, is MIT)
  - Bugfix: typographical errors

## Version 0.13 (19 July 2015)
This release fixes a critical bug and updates a few other non-functional
files.

  - Bugfix: was not in correct directory when removing package files
  - Update: Zsh completion understands reset command

## Version 0.12 (17 July 2015)
This marks the first release where gb is used to build the project. That means
that all the dependencies for this project (apart from gb itself) are contained
within the project. Other changes are:

  - New: simple shortcutting filter commands is now possible. Instead of
    `aur.new` you can write `a.n`. At the moment, both parts are required.
  - New: library `shortry` to implement some of shortcutting behavior.
  - New: filter `db.missing` shows packages in local database which do not have
    respective files. These are candidates for deletion.
  - New: filter command can negate filters.
  - New: default configuration is written if there is no configuration. If
    default configuration is not edited, repoctl refuses to run.

There are probably many more changes, but at the moment I can't be bothered to
hunt them all down.

## Version 0.11 (18 December 2014)

  - New: The configuration file learned the field `ignore_aur`, which affects
    status and filter commands.

## Version 0.10 (1 October 2014)
This release changes the versioning scheme to use semantic versioning. Since
we are still changing a lot of program details and functionality, we are in
an unstable state, hence the major version number being 0. We chose 10 as
the minor number just because.

Additionally, there are some changes and updates:

  - New: Repoctl learned the `filter` command, which can take certain criteria.
  - Change: The `status` command takes no arguments anymore.
  - Change: The configuration file is in the [TOML](https://github.com/toml-lang/toml)
    format now, and it is being read (but not created yet).
