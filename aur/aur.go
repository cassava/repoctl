// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package aur lets you query the Arch Linux User Repository (AUR).
package aur

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/goulash/pacman"
	"github.com/goulash/pacman/alpm"
)

// Packages is a slice of Package with several methods facilitating
// sorting, iterating, and converting to pacman.Packages.
type Packages []*Package

func (pkgs Packages) Len() int      { return len(pkgs) }
func (pkgs Packages) Swap(i, j int) { pkgs[i], pkgs[j] = pkgs[j], pkgs[i] }
func (pkgs Packages) Less(i, j int) bool {
	if pkgs[i].Name != pkgs[j].Name {
		return pkgs[i].Name < pkgs[j].Name
	}
	return alpm.VerCmp(pkgs[i].Version, pkgs[j].Version) == -1
}

// Pkgs returns the entire slice as pacman.Packages.
func (pkgs Packages) Pkgs() pacman.Packages {
	results := make(pacman.Packages, len(pkgs))
	for i, p := range pkgs {
		results[i] = p.Pkg()
	}
	return results
}

// Iterate calls f for each package in the list of packages.
func (pkgs Packages) Iterate(f func(pacman.AnyPackage)) {
	for _, p := range pkgs {
		f(p)
	}
}

// NotFoundError is returned when a package could not be found on AUR.
//
// The error message returned is different dependent on the number of packages
// that could not be found.
type NotFoundError struct {
	Names []string
}

func (e NotFoundError) Error() string {
	n := len(e.Names)
	if n == 1 {
		return fmt.Sprintf("package %q could not be found on AUR", e.Names[0])
	} else if n == 2 {
		return fmt.Sprintf("packages %q and %q could not be found on AUR", e.Names[0], e.Names[1])
	}

	// We have three or more packages that we could not find.
	var buf bytes.Buffer
	buf.WriteString("packages ")
	for _, nam := range e.Names[:n-1] {
		buf.WriteRune('"')
		buf.WriteString(nam)
		buf.WriteString(`", `)
	}
	buf.WriteString(`and "`)
	buf.WriteString(e.Names[n-1])
	buf.WriteString(`" could not be found on AUR`)
	return buf.String()
}

// IsNotFound returns true if the error is an instance of NotFoundError.
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

// response is what the AUR api returns:
//
//  {
//      "version":5,
//      "type":"multiinfo",
//      "resultcount":1,
//      "results": [
//          {
//              "ID":279276,
//              "Name":"telegram-desktop",
//              "PackageBaseID":95783,
//              "PackageBase":"telegram-desktop",
//              "Version":"0.9.28-1",
//              "Description":"Official desktop version of Telegram messaging app.",
//              "URL":"https:\/\/desktop.telegram.org\/",
//              "NumVotes":43,
//              "Popularity":8.94582,
//              "OutOfDate":null,
//              "Maintainer":"eduardosm",
//              "FirstSubmitted":1436478182,
//              "LastModified":1456515415,
//              "URLPath":"\/cgit\/aur.git\/snapshot\/telegram-desktop.tar.gz",
//              "Depends":["ffmpeg","icu","jasper","libexif","libmng","libwebp",
//                         "libxkbcommon-x11","libinput","libproxy","mtdev",
//                         "openal","libva","desktop-file-utils","gtk-update-icon-cache"],
//              "MakeDepends":["git","patch","libunity","libappindicator-gtk2"]
//              "License":["GPL3"]
//              "Keywords":[]
//          }
//      ]
//  }
type response struct {
	Version     int        `json:"version"`
	Type        string     `json:"type"`
	ResultCount int        `json:"resultcount"`
	Results     []*Package `json:"results"`
}

// Package is the information that we can retrieve about a package that is
// hosted on the Arch Linux User Repository (AUR), version 4.
//
// JSON Example:
//
//	{
//		"ID": 213309,
//		"Name": "repoctl",
//		"PackageBaseID": 96153,
//		"PackageBase": "repoctl",
//		"Version": "0.13-2",
//		"Description": "A supplement to repo-add and repo-remove which simplifies managing local repositories",
//		"URL": "https:\/\/github.com\/cassava\/repoctl",
//		"NumVotes": 1,
//		"OutOfDate": 0,
//		"Maintainer": "cassava",
//		"FirstSubmitted": 1437296687,
//		"LastModified": 1437298275,
//		"License": "MIT",
//		"URLPath": "\/cgit\/aur.git\/snapshot\/repoctl.tar.gz",
//		"CategoryID": 1,
//		"Popularity": 0
//	}
type Package struct {
	ID             uint64
	Name           string
	PackageBaseID  uint64
	PackageBase    string
	Version        string
	Description    string
	URL            string
	NumVotes       int
	Popularity     float64
	OutOfDate      int
	Maintainer     string
	FirstSubmitted uint64
	LastModified   uint64
	URLPath        string
	Depends        []string
	MakeDepends    []string
	License        []string
	Keywords       []string
}

// Pkg converts an aur.Package into a pacman.Package.
//
// Note that only a few fields in the resulting Package are actually filled in,
// namely Origin, Name, Version, Description, URL, and License. This is all the
// information that we are able to retrieve.
func (p *Package) Pkg() *pacman.Package {
	return &pacman.Package{
		Origin:      pacman.AUROrigin,
		Name:        p.Name,
		Base:        p.PackageBase,
		Version:     p.Version,
		Description: p.Description,
		URL:         p.URL,
		// TODO: License is string, but p.License is []string
		// Has there been a format change?
		//License:     p.License,
		Depends:     p.Depends,
		MakeDepends: p.MakeDepends,
	}
}

// PkgName returns the unique name of the package.
func (p *Package) PkgName() string { return p.Name }

// PkgVersion returns the version string of the package.
func (p *Package) PkgVersion() string { return p.Version }

// PkgDepends returns the dependencies of the package.
func (p *Package) PkgDepends() []string { return p.Depends }

// PkgMakeDepends returns the make dependenciess of the package.
func (p *Package) PkgMakeDepends() []string { return p.MakeDepends }

// DownloadURL returns the URL for downloading the PKGBUILD tarball.
func (p *Package) DownloadURL() string {
	return fmt.Sprintf("https://aur.archlinux.org%s", p.URLPath)
}

const (
	searchURL    = "https://aur.archlinux.org/rpc.php?v=5&type=search&by=name&arg=%s"
	multiInfoURL = "https://aur.archlinux.org/rpc.php?v=5&type=multiinfo&arg[]=%s"
	multiInfoArg = "&arg[]="
)

// generateMultiInfoURL creates a URL that gets the package information from AUR.
func generateMultiInfoURL(args []string) string {
	na := make([]string, len(args))
	for i, s := range args {
		na[i] = url.QueryEscape(s)
	}
	return fmt.Sprintf(multiInfoURL, strings.Join(na, multiInfoArg))
}

func SearchByName(query string) (Packages, error) {
	q := fmt.Sprintf(searchURL, url.QueryEscape(query))
	resp, err := http.Get(q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msg response
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&msg)
	return msg.Results, err
}

// Read reads package information from the Arch Linux User Repository (AUR)
// online.
//
// If a package cannot be found, (nil, *NotFoundError) is returned.
func Read(pkgname string) (*Package, error) {
	q := generateMultiInfoURL([]string{pkgname})
	resp, err := http.Get(q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msg response
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	if msg.ResultCount == 0 {
		return nil, &NotFoundError{Names: []string{pkgname}}
	}
	return msg.Results[0], nil
}

// ReadAll reads multiple packages from the Arch Linux User Repository (AUR)
// at once.
//
// If any packages cannot be found, (Packages, *NotFoundError) is returned.
// That is, all successfully read packages are returned.
func ReadAll(pkgnames []string) (Packages, error) {
	// We only query at most 200 packages at a time, the limit currently
	// appears to be 250, but we'll stay well beneath that for now.
	const limit = 200
	if len(pkgnames) <= limit {
		return readAll(pkgnames)
	}

	var pkgs Packages
	var err *NotFoundError
	for len(pkgnames) > 0 {
		// Select next slice of messages
		var slice []string
		if len(pkgnames) > limit {
			slice = pkgnames[:limit]
			pkgnames = pkgnames[limit:]
		} else {
			slice = pkgnames
			pkgnames = []string{}
		}

		// Query selected slice of messages
		p, e := readAll(slice)
		if e != nil {
			nfe, ok := e.(*NotFoundError)
			if !ok {
				return nil, e
			}

			if err == nil {
				err = nfe
			} else {
				err.Names = append(err.Names, nfe.Names...)
			}
		}
		pkgs = append(pkgs, p...)
	}
	return pkgs, err
}

func readAll(pkgnames []string) (Packages, error) {
	q := generateMultiInfoURL(pkgnames)
	resp, err := http.Get(q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msg response
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&msg)
	if err != nil {
		return nil, err
	}
	if msg.ResultCount != len(pkgnames) {
		m := make(map[string]bool)
		nfe := &NotFoundError{
			Names: make([]string, 0, len(pkgnames)-msg.ResultCount),
		}
		for _, i := range msg.Results {
			m[i.Name] = true
		}
		for _, s := range pkgnames {
			if !m[s] {
				nfe.Names = append(nfe.Names, s)
			}
		}
		return msg.Results, nfe
	}
	return msg.Results, nil
}
