// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
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
	OutOfDate      int
	Maintainer     string
	FirstSubmitted uint64
	LastModified   uint64
	License        string
	URLPath        string
}

func generateURL(arg string) string {
	queryURL := "https://aur.archlinux.org/rpc.php?type=info&arg=%s"
	return fmt.Sprintf(queryURL, url.QueryEscape(arg))
}

// ReadAUR reads package information from the Arch Linux User Repository
// online.  If the package cannot be found, but no unexpected error occurred,
// then (nil, nil) is returned.
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
	if err != nil || msg.ResultCount != 1 {
		// If there is an error at this stage, treat it like 'not found'.
		return nil, nil
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

// ConcurrentlyReadAUR reads the pkgnames using n goroutines.
//
// Errors are passed through the channel if the channel is not nil, otherwise
// they are ignored. Make sure you handle the errors right away, like so:
//
//	ch := make(chan error)
//	go func() {
//		for err := range ch {
//			fmt.Println("error:", err)
//		}
//	}()
//	pkgs := pacman.ConcurrentlyReadAUR(pkgs, n, ch)
//	close(ch)
//
// Because if you don't, the program will probably run into a deadlock when
// there is an error. Note that ConcurrentlyReadAUR does not close the channel,
// you have to do that yourself.
func ConcurrentlyReadAUR(pkgnames []string, n int, ch chan<- error) map[string]*Package {
	var wg sync.WaitGroup
	var mu sync.Mutex
	jobs := make(chan string)
	pkgs := make(map[string]*Package, len(pkgnames))

	jobber := func() {
		for j := range jobs {
			p, err := ReadAUR(j)
			if err != nil {
				if ch != nil {
					ch <- err
				}
				continue
			}

			mu.Lock()
			pkgs[j] = p
			mu.Unlock()
		}
		wg.Done()
	}

	// Start n concurrent job takers
	for i := 0; i < n; i++ {
		wg.Add(1)
		go jobber()
	}

	// Start a producer of jobs
	go func() {
		for _, q := range pkgnames {
			jobs <- q
		}
		close(jobs)
	}()

	wg.Wait()
	return pkgs
}
