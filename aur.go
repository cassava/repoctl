// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type aurResponse struct {
	ResultCount int
	Results     aurPkgInfo
}

type aurPkgInfo struct {
	ID             uint64
	Name           string
	Version        string
	Description    string
	URL            string
	NumVotes       int
	OutofDate      bool
	Maintainer     string
	FirstSubmitted time.Time
	LastModified   time.Time
	License        string
	URLPath        string
}

func generateURL(arg string) string {
	queryURL := "https://aur.archlinux.org/rpc.php?type=info&arg=%s"
	return fmt.Sprintf(queryURL, url.QueryEscape(arg))
}

// ReadAUR reads package information from the Arch Linux User Repository online.
//
// Note that only a few fields in the resulting Package are actually filled in,
// namely Origin, Name, Version, Description, URL, and License. This is all the
// information that we are able to retrieve.
func ReadAUR(pkgname string) (*Package, error) {
	q := generateURL(pkgname)
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

	if msg.ResultCount != 1 {
		if msg.ResultCount > 1 {
			return nil, errors.New("unexpected: too many packages in resultset")
		}
		return nil, errors.New("package not found")
	}
	info := msg.Results
	pkg := Package{
		Origin:      AUROrigin,
		Name:        info.Name,
		Version:     info.Version,
		Description: info.Description,
		URL:         info.URL,
		License:     info.License,
	}

	return &pkg, nil
}
