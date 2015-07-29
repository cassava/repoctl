// Copyright (c) 2014-2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

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

type aurResponse struct {
	ResultCount int
	Results     []*AURPackage
}

// AURPackage is the information that we can retrieve about a package that is
// hosted on the Arch Linux User Repository (AUR), version 4.
//
// I am not entirely sure what the difference is between (ID, Name) and
// (PackageBaseID, PackageBase).
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
type AURPackage struct {
	ID             uint64
	Name           string
	PackageBaseID  uint64
	PackageBase    string
	Version        string
	Description    string
	URL            string
	NumVotes       int
	OutOfDate      bool `json:",int"`
	Maintainer     string
	FirstSubmitted uint64
	LastModified   uint64
	License        string
	URLPath        string
	CategoryID     int
	Popularity     float64
}

// Package converts an AURPackage into a Package.
//
// Note that only a few fields in the resulting Package are actually filled in,
// namely Origin, Name, Version, Description, URL, and License. This is all the
// information that we are able to retrieve.
func (ap *AURPackage) Package() *Package {
	return &Package{
		Origin:      AUROrigin,
		Name:        ap.Name,
		Version:     ap.Version,
		Description: ap.Description,
		URL:         ap.URL,
		License:     ap.License,
	}
}

// DownloadURL returns the URL for downloading the PKGBUILD tarball.
func (ap *AURPackage) DownloadURL() string {
	return fmt.Sprintf("https://aur.archlinux.org%s", ap.URLPath)
}

const (
	apiURL = "https://aur4.archlinux.org/rpc.php?type=multiinfo&arg[]=%s"
	apiArg = "&arg[]="
)

// generateURL creates a URL that gets the package information from AUR.
func generateURL(args []string) string {
	na := make([]string, len(args))
	for i, s := range args {
		na[i] = url.QueryEscape(s)
	}
	return fmt.Sprintf(apiURL, strings.Join(na, apiArg))
}

// ReadAUR reads package information from the Arch Linux User Repository (AUR)
// online.
//
// If a package cannot be found, (nil, *NotFoundError) is returned.
func ReadAUR(pkgname string) (*AURPackage, error) {
	q := generateURL([]string{pkgname})
	resp, err := http.Get(q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msg aurResponse
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

// ReadAllAUR reads multiple packages from the Arch Linux User Repository (AUR)
// at once.
//
// If any packages cannot be found, ([]*AURPackage, *NotFoundError) is returned.
// That is, all successfully read packages are returned.
func ReadAllAUR(pkgnames []string) ([]*AURPackage, error) {
	q := generateURL(pkgnames)
	resp, err := http.Get(q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msg aurResponse
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
