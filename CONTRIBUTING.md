Contributor's Guide
===================

I don't yet have any specific requirements for new contributors
other than the ones that Github implicitely provides.

This document acts as a cookbook for how to do some specific (mundane) tasks.

## Updating Dependencies

Make sure the working tree is clean before starting this.
Then, basically run the following commands:

    go get -u ./...
    go mod tidy
    go mod vendor

And then build and test to ensure everything is OK:

    go build ./...
    go test ./...

Ensure that the CI is using the same version of Go as is defined in `go.mod`:

    $EDITOR .github/workflows/ci.yaml

Create a new commit with all the changes:

    git add .
    git commit -m "Update dependencies"

## Creating a Release

 1. Ensure that the modules have been updated.

 2. Update `version.go`, especially for these metadata items:

    - version (X.Y.Z)
    - date (YYYY-MM-DD)
    - copyright (update to include current year if necessary)

 3. Add a new entry to `NEWS.md`.
    (The same text will be used later when we create a Github release.)

 4. Commit these changes:

        git add .
        git commit -m "Release version X.Y.Z"

    However, do not push this yet. The working directory should be clean.

 5. Create an archive for testing.

        git archive --prefix=repoctl-X.Y.Z/ -o ../repoctl.aur/repoctl-X.Y.Z.tar.gz HEAD

    Here we refer to `repoctl.aur` for the first time. This is just the
    [repository](https://aur.archlinux.org/packages/repoctl) that contains `PKGBUILD` for AUR.
    If you have the rights for pushing to this, then good, otherwise you
    will need to contact the owner of that package on AUR.

 6. In the AUR directory, you need to update the `PKGBUILD` file, in particular
    for the new version:

    - version (set to X.Y.Z)
    - release (reset to 1)
    - md5sum (update for source archive)

    You can get the checksum with: 

        md5sum repoctl-X.Y.Z.tar.gz

    When this is ready, you can update `.SRCINFO`:

        makepkg --printsrcinfo > .SRCINFO

    There should be two files modified: `PKGBUILD` and `.SRCINFO`.
    Now we can build the package and check that it works:

        makepkg
        sudo pacman -U repoctl-X.Y.Z-1.pkg.tar.zst

    Do some rudimentary tests to ensure that everything is working
    as expected:

        which repoctl    # should be /usr/bin/repoctl
        repoctl version  # should match current date and version

    If everything is OK up until this point, then commit the changes:

        git commit -a -m "Update repoctl to version X.Y.Z"

    But hold off on pushing.

 7. In the main repository, create a tag, then push HEAD and tags:

        git tag -a vX.Y.Z -m "repoctl version X.Y.Z release"
        git push
        git push --tags

 8. In the Github Releases page, draft a new release.

    - Copy the notes from `NEWS.md` into the description.
    - Upload the source archive that we created in step 6.
    - Create the release.

 9. Back in the AUR repository for repoctl, clean the working tree and
    ensure that everything can be built cleanly from scratch:

        git clean -xdf
        makepkg

    If this works out fine, then push the commit to update the AUR entry:

        git push

And that is how a release is made.
