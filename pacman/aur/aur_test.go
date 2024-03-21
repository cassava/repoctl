// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package aur

import "testing"

var (
	exists    = []string{"repoctl", "fairsplit", "moped"}
	notExists = []string{"repoctl-34534", "arstaorsf", "911222234"}
	invalid   = []string{"-", "_", "*", "-1q"}
	many      = []string{
		"a2ps",
		"a52dec",
		"aalib",
		"abcde",
		"abs",
		"abuse",
		"acl",
		"acpi",
		"acpid",
		"acroread",
		"adobe-source-code-pro-fonts",
		"adwaita-icon-theme",
		"aircrack-ng",
		"akonadi",
		"akonadi-contacts",
		"alacritty-git",
		"alex",
		"alsa-lib",
		"alsa-plugins",
		"android-sdk-platform-tools",
		"apache",
		"apache-ant",
		"apm",
		"apr",
		"apr-util",
		"aqbanking",
		"archlinux-keyring",
		"ardour",
		"arpack",
		"asciidoc",
		"aspell",
		"aspell-de",
		"aspell-en",
		"asunder",
		"at",
		"at-spi2-atk",
		"at-spi2-core",
		"atk",
		"atkmm",
		"atom",
		"attica-qt4",
		"attica-qt5",
		"attr",
		"aubio",
		"audacity",
		"audiofile",
		"autoconf",
		"autoconf-archive",
		"automake",
		"avahi",
		"avidemux-cli",
		"avidemux-qt",
		"awesome-git",
		"awmtt",
		"aws-cli",
		"babl",
		"baloo",
		"baloo4-akonadi",
		"bash",
		"batterymon-clone",
		"bc",
		"biber",
		"bind-tools",
		"binutils",
		"bison",
		"blas",
		"bless",
		"bluez",
		"bluez-cups",
		"bluez-firmware",
		"bluez-libs",
		"bluez-plugins",
		"bluez-tools",
		"bluez-utils",
		"boost",
		"boost-libs",
		"brasero",
		"bridge-utils",
		"bsdiff",
		"btrfs-progs",
		"bubblewrap",
		"bzip2",
		"bzr",
		"c++utilities",
		"c-ares",
		"ca-certificates",
		"ca-certificates-cacert",
		"ca-certificates-mozilla",
		"ca-certificates-utils",
		"cabal-install",
		"cabextract",
		"cairo",
		"cairo-perl",
		"cairomm",
		"calc",
		"calibre",
		"calligra",
		"cantata-git",
	}
)

func TestRead1(z *testing.T) {
	for _, n := range exists {
		i, err := Read(n)
		if err != nil {
			z.Errorf("unexpected error: %s", err)
		}
		if i == nil {
			z.Errorf("expected i = non-nil, got nil")
		} else if i.Name != n {
			z.Errorf("wrong name returned")
		}
	}
}

func TestRead2(z *testing.T) {
	for _, n := range notExists {
		i, err := Read(n)
		if i != nil {
			z.Errorf("expecting i to be nil")
		}
		if err == nil {
			z.Errorf("expecting error, got nil")
		} else if nf, ok := err.(*NotFoundError); ok {
			if len(nf.Names) != 1 {
				z.Errorf("wrong number of names returned")
			} else if nf.Names[0] != n {
				z.Errorf("wrong name returned")
			}
		} else {
			z.Errorf("unexpected error: %s", err)
		}
	}
}

func TestReadAll1(z *testing.T) {
	is, err := ReadAll(exists)
	if err != nil {
		z.Errorf("unexpected error: %s", err)
	}
	if is == nil {
		z.Errorf("expected i = non-nil, got nil")
	}
}

func TestReadAll2(z *testing.T) {
	is, err := ReadAll(notExists)
	if len(is) != 0 {
		z.Errorf("expecting is to have zero elements")
	}
	if err == nil {
		z.Errorf("expecting error, got nil")
	} else if nf, ok := err.(*NotFoundError); ok {
		if len(nf.Names) != len(notExists) {
			z.Errorf("wrong number of names returned")
		} else {
			for i, n := range notExists {
				if nf.Names[i] != n {
					z.Errorf("wrong name returned")
				}
			}
		}
	} else {
		z.Errorf("unexpected error: %s", err)
	}
}

func TestReadMany(z *testing.T) {
	is, err := ReadAll(many)
	if len(is) == 0 {
		z.Errorf("expecting to have more than zero elements")
	}
	if err == nil {
		z.Errorf("expecting error, got nil")
	} else if nf, ok := err.(*NotFoundError); ok {
	next_package:
		for _, n := range many {
			// Either the package was found:
			for _, p := range is {
				if p.Name == n {
					continue next_package
				}
			}

			// Or the package was not found:
			for _, p := range nf.Names {
				if p == n {
					continue next_package
				}
			}

			// Otherwise something is wrong!
			z.Errorf("expected package %s to be found or not found", n)
		}
	} else {
		z.Errorf("unexpected error: %s", err)
	}
}

func TestDownloadURL(z *testing.T) {
	i, err := Read("repoctl")
	if err != nil {
		z.Errorf("unexpected error: %s", err)
		z.FailNow()
	}
	if i.DownloadURL() != "https://aur.archlinux.org/cgit/aur.git/snapshot/repoctl.tar.gz" {
		z.Errorf("download url incorrect: %s", i.DownloadURL())
	}
}

func TestDownloadURLWithBaseName(z *testing.T) {
	i, err := Read("transgui-qt")
	if err != nil {
		z.Errorf("unexpected error: %s", err)
		z.FailNow()
	}
	if i.DownloadURL() != "https://aur.archlinux.org/cgit/aur.git/snapshot/transgui.tar.gz" {
		z.Errorf("download url incorrect: %s", i.DownloadURL())
	}
}
