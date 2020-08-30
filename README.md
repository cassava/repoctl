repoctl
========

The **repoctl** program helps you manage a local repository of Pacman
packages (as found on Arch Linux and derivatives).

In the following video you can see a whirlwind tour of what repoctl can do for
someone who is just starting out with local repositories: We will search for an
extension for the `pass` tool and add it to a new local repository.

[![asciicast](https://asciinema.org/a/yZFlXwwax8hjLYF0R0JXX6GWP.svg)](https://asciinema.org/a/yZFlXwwax8hjLYF0R0JXX6GWP)

This isn't all repoctl can do; veterans might find they use the `status` and
`update` commands more for day-to-day managing of local repositories.

A look at the available commands may help give an overview:

 - **add** - Copy and add packages to the repository.
 - **conf** - Create, edit, or show the repoctl configuration.
 - **down** - Download and extract tarballs from AUR.
 - **host** - Host repository on a network.
 - **list** - List packages that belong to the managed repository.
 - **query** - Query package information from AUR.
 - **remove** - Remove and delete packages from the database.
 - **reset** - (Re-)create repository database.
 - **search** - Search for packages on AUR.
 - **status** - Show pending changes and packages that can be upgraded.
 - **update** - Update database in repository to match filesystem.

See the [NEWS](NEWS.md) for the latest changes in repoctl!

### Installation

The recommended method is to install the `repoctl` package from [AUR](https://aur.archlinux.org/packages/repoctl),
as this package installs other useful files, such as the completion scripts.

Alternatively, if you have [Go](https://golang.org) installed:
```sh
git clone https://github.com/cassava/repoctl.git
cd repoctl
# Install repoctl to $GOPATH/bin, or specify -o OUTPUT
go install
# Get help on where to install completion files to:
repoctl completion --help
```

You may want to switch to the `devel` branch if you want the bleeding edge.

### Basic Usage

Before you can use *really* use repoctl, you need to create a configuration file,
but there's a lot you can do without any configuring.

1. Search AUR:
   ```console
   $ repoctl search tomb-
   aur/mediatomb-git 7ab7616-1 (2)
       Free UPnP/DLNA media server mediatomb
   aur/gtomb-git 0.7.1-3 (2)
       GUI wrapper for Tomb, the crypto undertaker
   aur/tomb-git 2.6.r7.g6f2ce59-1 (6)
       simple tool to manage encrypted storage
   aur/tomb-kdf-git 2.6.r7.g6f2ce59-1 (6)
       Crypto Undertaker extensions to improve password security
   aur/mediatomb-samsung-tv 0.12.1-12 (8)
       Free UPnP/DLNA media server with Samsung TV compatibility
   aur/tomb-kdf 2.7-2 (45)
       Crypto Undertaker extensions to improve password security
   ```

2. Query specific packages on AUR:
   ```console
   $ repoctl query tomb
   aur/tomb 2.7-2 (45)
       Name: tomb
       Version: 2.7-2
       Description: Crypto Undertaker, a simple tool to manage encrypted storage
       URL: https://www.dyne.org/software/tomb/
       Licenses: GPL3
       Dependencies: bc cryptsetup gnupg sudo zsh e2fsprogs inetutils
       Optional Dependencies:
           steghide
           dcfldd
           qrencode
           swish-e
       Snapshot URL: https://aur.archlinux.org/cgit/aur.git/snapshot/tomb.tar.gz
       Maintainer: parazyd
       Votes: 45
       Popularity: 0.355983
       First Submitted: 2011-04-15 17:20:00 +0200 CEST
       Last Updated: 2020-01-03 13:57:47 +0100 CET
       Out-Of-Date: false
   ```

3. Download packages from AUR, including their dependencies:
   ```console
   $ repoctl down -r pass-tomb
   Downloading: pass-tomb
   Downloading: tomb
   ```

### Configuration

Before we can actually start managing a local repository, repoctl needs to
know where it is. No one really enjoys working with configuration files,
so repoctl will help you out a little here.

1. Create a new configuration, with our repo in `~/pkgs`:
   ```console
   $ repoctl conf new ~/pkgs/sirius.db.tar.zst
   Writing new configuration file at: /home/you/.config/repoctl/config.toml
   ```

2. Initialize the repository:
   ```console
   $ repoctl reset
   Creating database: /home/ben/pkgs/sirius.db.tar.zst
   ```

Now you should be set to start adding packages to your repository.

If you want to fine-tune the configuration values or just see what's there,
repoctl will show you your configuration (`repoctl conf show`) as well as
launch you into it with your favorite editor (`repoctl conf edit`), as set in
the environment variable `EDITOR`.

3. Inspect your configuration.
   ```console
   $ repoctl conf show
   Current configuration:
       columnate = false
       color = "auto"
       quiet = false

       current_profile = ""
       default_profile = "default"

       [profiles.default]
           repo = "/home/you/pkgs/sirius.db.tar.zst"
           add_params = []
           rm_params = []
           ignore_aur = []
           require_signature = false
           backup = false
           backup_dir = ""
           interactive = false
           pre_action = ""
           post_action = ""
   ```

4. Edit your configuration:
   ```console
   $ repoctl conf edit
   ```

### Managing Your Repository

Now, we can add and manipulate packages in the specified local repository.

1. Add packages to the repository:
   ```console
   $ repoctl add <tab>
   $ repoctl add pass-extension-tail-1.2.0-1-any.pkg.tar.zst
   Copying and adding to repository: pass-extension-tail-1.2.0-1-any.pkg.tar.zst
   Adding package to database: /home/you/pkgs/pass-extension-tail-1.2.0-1-any.pkg.tar.zst
   ```
   If you installed the completion, you really should take advantage of it,
   unless of course you are automating the procedure.

2. Remove packages from the repository:
   ```console
   $ repoctl rm <tab>
   $ repoctl rm pass-extension-tail
   Removing package from database: pass-extension-tail
   Deleting: pass-extension-tail-1.2.0-1-any.pkg.tar.zst
   ```
   Yes, package names from your repository are also completed.

### Managing Updates To Your Packages

Of course, the initial compilation and adding of packages isn't the trouble,
it's keeping them all up-to-date. This is what repoctl was originally made
for: to tell me which packages have been updated on AUR and get them for me.

1. Show which packages have updates on AUR:
   ```console
   $ repoctl status -a
   On repo sirius

       krop: upgrade(0.4.11-1 -> 0.6.0-1)
       spotify: upgrade(1.0.98.78-1 -> 1:1.1.10.546-4)
       tmuxinator: upgrade(0.8.1-1 -> 2.0.1-1)
       ttf-ms-win10: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-japanese: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-korean: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-other: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-sea: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-thai: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-zh_cn: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
       ttf-ms-win10-zh_tw: upgrade(10.0.14393-3 -> 10.0.18362.116-2)
   ```
   You can't see it here, but this is all nicely colored in your terminal.

2. Download all updated packages:
   ```console
   $ repoctl down -u -o build-order.txt
   Downloading: tmuxinator
   Downloading: krop
   Downloading: python-poppler-qt5
   Downloading: ttf-ms-win10
   Downloading: ruby-xdg
   Downloading: ruby-erubis

   $ cat build-order.txt
   ruby-erubis
   ruby-xdg
   ttf-ms-win10
   python-poppler-qt5
   krop
   tmuxinator
   ```

What's the `build-order.txt` file for, you say? I'm glad you asked. Some
packages, such as `tmuxinator` up there, have dependencies on other packages
(in this case, `ruby-xdg` and `ruby-erubis`). If these packages are in AUR,
then we need to fetch them too. This is what the `-r` (`--recursive`) flag is
good for, and if we specify the `-o` flag (`--order`) it is implied.

We can use this list to our advantage, and with some Bash fu compile the
whole lot of packages and add them to the repository:
```bash
#!/bin/bash
set -e
repoctl down -u -o build-order.txt
for pkg in $(cat build-order.txt); do
    (
        cd "$pkg"
        makepkg -cs
        repoctl add *.pkg.tar.zst
        cd ..
        rm -rf "$pkg"
    )
done
```

### Tips and Tricks

1. Using `PKGDEST` in `/etc/makepkg.conf`

   You can configure makepkg to put all generated packages into a directory
   of your choosing. If you want, you can set `PKGDEST` to your repository
   directory, and then just run `repoctl update` to do the rest.

2. Auto-completion for everything!

   Since version 0.21, auto-completion depends strongly on the repoctl tool
   itself. This lets us do some pretty wild things, like query AUR, read
   your configuration, or even read the repository database specified in
   the profile you just added to the command-line invocation.

   Make sure you install the completions for your shell if you haven't done so
   yet. There is a hidden command for exporting the shell completion:
   ```
   $ repoctl completion
   ...
   ```

3. Configuring multiple repositories

   Configuration profiles are supported since version 0.21. These let you
   have more than one profile that you can then choose at runtime.

   The important configuration settings are:
   ```
   default_profile = "default"

   [profiles.default]
     repo = "/home/you/pkgs/sirius.db.tar.zst"

   [profiles.release]
     repo = "/home/you/public/pkgs/sirius-release.db.tar.zst"
     require_signature = true
     backup = true
     backup_dir = "backup/"
   ```
   See the `conf` command for more information on this.

4. Migrating your configuration file

   The configuration file has changed significantly since version 0.21 in
   order to support profiles. This means that some configuration values
   are deprecated and no longer supported and the format in general is
   different.

   Fear not! Not only is your old configuration auto-migrated, but you
   can make this migration permanent with the `conf migrate` command:
   ```console
   $ repo conf migrate
   Backing up current configuration to: /home/ben/.config/repoctl/config.toml.bak.3
   Writing new configuration file to: /home/ben/.config/repoctl/config.toml
   ```

5. Caching obsolete packages

   Sometimes you might want to hold on to the obsolete packages and leave them
   in the directory at the same time, and use a tool like [paccache](https://wiki.archlinux.org/index.php/pacman)
   to manage them. You can easily enable this in your config:
   ```toml
   [profiles.default]
       backup = true
       backup_dir = ""
   ```
   Now, obsolete packages will be ignored. They will also be ignored when
   removing packages from the database.

6. Packages on a remote filesystem

   If you have a super fast internet connection and want your packages on a remote
   server, you can *try* to get repoctl to play along with the `pre_action` and
   `post_action` options in the configuration file:
   ```toml
   [profiles.default]
       pre_action = "sshfs server:location ~/localmnt"
       post_action = "fusermount -u ~/localmnt"
   ```
   This will definitely break auto-completion and if some error happens, the
   `post_action` might not be executed, so **I don't recommend this**.
   Instead, it's much better to simply rsync your packages at the end.

### Getting Help

These are not the only things that repoctl can do, to get a fuller picture,
have a look at the help, which you can always get by using the `--help` flag or
by running:
```console
$ repoctl help [cmd]
[...]
```

Chances are good you might encounter errors or have a bright idea about how
to improve repoctl. If you do, I would love to hear about it!

Have a look at the existing issues or create a new issue at [GitHub](https://github.com/cassava/repoctl/issues).

Enjoy!
