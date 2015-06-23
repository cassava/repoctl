Repoctl Releases
================

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
