// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cassava/repoctl/pacman/aur"
	"github.com/goulash/pr"
	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(queryCmd)
}

var queryCmd = &cobra.Command{
	Use:   "query [pkgname...]",
	Short: "query package information from AUR",
	Long: `Query package information from AUR.

  This command queries AUR for the specified packages and returns as much
  information on these packages as AUR gives us. The results are combined and
  sorted alphabetically.

  Note that this command is very similar to the results given from "search -i"
  command, but it uses a different AUR request. This command shows the
  following additional metadata:

    - Groups
    - Dependencies
    - Make Dependencies
    - Optional Dependencies
    - Conflicts
    - Provides
    - Replaces
    - Keywords

  Metadata properties that are empty are not shown.
`,
	Example: `  repoctl query firefox56 flirc-bin`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		pkgs, err := aur.ReadAll(args)
		if err != nil {
			nfe, ok := err.(*aur.NotFoundError)
			if !ok {
				return err
			}
			for _, n := range nfe.Names {
				fmt.Fprintf(os.Stderr, "warning: unknown package %s\n", n)
			}
		}

		// Get the terminal width and fallback to a massive value if it's not
		// available. This prevents wrapping and lets us for example grep the
		// output better.
		terminalWidth := pr.StdoutTerminalWidth()
		if terminalWidth <= 0 {
			// FIXME: This is a hack
			terminalWidth = 1024
		}

		// Print the list
		pkgset := make(map[string]bool)
		for _, p := range pkgs {
			// Only add unique names to the list of packages
			if pkgset[p.Name] {
				continue
			}
			pkgset[p.Name] = true

			col.Printf("@{!m}aur/@{!w}%s @{!g}%s @{r}(%d)\n@|", p.Name, p.Version, p.NumVotes)
			col.Printf("@.%s\n", formatAURPackageInfo(p, terminalWidth))
		}

		return nil
	},
}

func formatAURPackageInfo(p *aur.Package, hspace int) string {
	// We want formatList to give us something like this:
	//		Depends: package package package package package
	//				 package package package package
	wrap := func(xs []string, prefixLen int) string {
		var buf strings.Builder

		n := prefixLen
		for i := 0; i < len(xs); i++ {
			x := xs[i]
			k := len(x)
			if n+k+1 > hspace && n != prefixLen {
				// If n == prefixLen, then that means we are at the beginning of the line,
				// and we still don't have enough space. We'll just have to deal with it.
				// A possible optimization here would be to try to reduce prefixLen and
				// see if it would fit then. But that can be done some other day.

				// Add a newline and the prefix
				buf.WriteRune('\n')
				buf.WriteString(strings.Repeat(" ", prefixLen))
				n = prefixLen
			}
			if n != prefixLen {
				buf.WriteRune(' ')
			}
			buf.WriteString(x)
			n += k + 1
		}

		return buf.String()
	}

	var buf strings.Builder
	fmt.Fprintf(&buf, "    Name: %s\n", p.Name)
	if p.PackageBase != p.Name {
		fmt.Fprintf(&buf, "    Base Name: %s\n", p.PackageBase)
	}
	fmt.Fprintf(&buf, "    Version: %s\n", p.Version)
	fmt.Fprintf(&buf, "    Description: %s\n", wrap(strings.Split(p.Description, " "), 17))
	if len(p.URL) != 0 {
		fmt.Fprintf(&buf, "    URL: %s\n", p.URL)
	}

	// The following is not available from the information we get when using the
	// search API, but might be useful in the future.
	//
	if len(p.License) > 0 {
		fmt.Fprintf(&buf, "    Licenses: %s\n", wrap(p.License, 14))
	}
	if len(p.Groups) > 0 {
		fmt.Fprintf(&buf, "    Groups: %s\n", wrap(p.Groups, 14))
	}
	if len(p.Provides) > 0 {
		fmt.Fprintf(&buf, "    Provides: %s\n", wrap(p.Provides, 14))
	}
	if len(p.Conflicts) > 0 {
		fmt.Fprintf(&buf, "    Conflicts: %s\n", wrap(p.Conflicts, 15))
	}
	if len(p.Replaces) > 0 {
		fmt.Fprintf(&buf, "    Replaces: %s\n", wrap(p.Replaces, 14))
	}
	if len(p.Depends) > 0 {
		fmt.Fprintf(&buf, "    Dependencies: %s\n", wrap(p.Depends, 16))
	}
	if len(p.OptDepends) > 0 {
		fmt.Fprintf(&buf, "    Optional Dependencies:\n")
		for _, d := range p.OptDepends {
			fmt.Fprintf(&buf, "        %s\n", d)
		}
	}
	if len(p.MakeDepends) > 0 {
		fmt.Fprintf(&buf, "    Build Dependencies: %s\n", wrap(p.MakeDepends, 15))
	}
	if len(p.Keywords) > 0 {
		fmt.Fprintf(&buf, "    Keywords: %s\n", wrap(p.Keywords, 14))
	}

	fmt.Fprintf(&buf, "    Snapshot URL: %s\n", p.URLPath)
	fmt.Fprintf(&buf, "    Maintainer: %s\n", p.Maintainer)
	fmt.Fprintf(&buf, "    Votes: %d\n", p.NumVotes)
	fmt.Fprintf(&buf, "    Popularity: %f\n", p.Popularity)
	fmt.Fprintf(&buf, "    First Submitted: %s\n", time.Unix(int64(p.FirstSubmitted), 0))
	fmt.Fprintf(&buf, "    Last Updated: %s\n", time.Unix(int64(p.LastModified), 0))
	fmt.Fprintf(&buf, "    Out-Of-Date: %v", p.OutOfDate != 0)
	return buf.String()
}
