#!/bin/bash
#

# Set the colors for the terminal
cBOLD="\e[1m"
cESC="\e[0m"

cBLUE="\e[1;34m"
cRED="\e[1;31m"
cGREEN="\e[1;32m"
cWHITE="\e[1;37m"
cYELLOW="\e[1;33m"
cCYAN="\e[1;36m"
cMAGENTA="\e[1;35m"

function infof() {
    local format="$1"
    shift 1
    printf "$cBOLD$cGREEN::$cWHITE $format$cESC\n" "$*"
}

function inputf() {
    local format="$1"
    shift 1
    printf "$cBOLD$cBLUE<<$cWHITE $format$cESC" "$*"
}

function errorf() {
    local format="$1"
    shift 1
    printf "$cBOLD$cRED!! Error:$cWHITE $format$cESC\n" "$*"
}

function warnf() {
    local format="$1"
    shift 1
    printf "$cBOLD$cYELLOW// Warning:$cWHITE $format$cESC\n" "$*"
}

# -------------------------------------------------------------------------- #

release_version=$1
release_date=$(date +"%d %B, %Y")
previous_tag=$(git describe --abbrev=0)
package_dir="../repoctl.aur"
main_branch="master"
github_url="https://github.com/cassava/repoctl"

confirm() {
    local prompt="$1"
    inputf "$prompt"
    printf " [Y/n] "
    local answer
    read answer
    [[ "$answer" == "y" || "$answer" == "Y" || -z "$answer" ]]
}

countdown() {
    local message="$1"
    local seconds=$2
    while [[ $seconds -ne 0 ]]; do
        printf "\r%s [%s]" "$message" "$seconds"
        seconds=$((seconds - 1))
        sleep 1
    done
    printf "\r%s           \n" "$message"
}

launch_editor() {
    local file="$1"
    local seconds=$2
    countdown "-> Launching ${EDITOR}" "$seconds"
    ${EDITOR} "$file"
}

# Ensure working directory of repo is clean.
if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
    warnf "Working directory is dirty!"
    echo
    git status --porcelain | sed -r 's/^/\t/'
    echo
    confirm "Do you want to continue?" || exit 2
fi

# Get version to release interactively if necessary.
if [[ -z "$release_version" ]]; then
    echo "Previous version tag: $previous_tag"
    inputf "Version to release: "
    read release_version
    if [[ -z "$release_version" ]]; then
        echo "Aborted."
        exit 1
    fi
fi

if [[ ! -f "$package_dir/PKGBUILD" ]]; then
    warnf "AUR package directory invalid: $package_dir"
    exit 1
fi

# Confirm release information.
infof "Verify release metadata"
echo "Git describe: $(git describe)"
echo "Previous version tag: $previous_tag"
echo "Release version: $release_version"
echo "Release date: $release_date"
confirm "Is this correct?" || exit 2

# Step 1:
if confirm "Update vendored dependencies?"; then
    (
        set -e
        go get -u ./...
        go mod tidy
        go mod vendor
        go build ./...
        go build .
        go test ./...
        go test .

        # Update .github/workflows/ci.yaml
        git add vendor go.mod go.sum
        git commit -m "Update dependencies"
    )
fi

# Step 2:
copyright_date="2016-$(date +"%Y")"
infof "Update version.go"
sed -r \
    -e "s/^(\t+Version: *\")(.*)(\",)$/\1${release_version}\3/" \
    -e "s/^(\t+Date: *\")(.*)(\",)$/\1${release_date}\3/" \
    -e "s/^(\t+Copyright: *\")(.*)(\",)/\1${copyright_date}\3/" \
    -i version.go
countdown "-> Showing git diff version.go" 2
git diff --color=always version.go
if ! confirm "Is version.go correct?"; then
    launch_editor version.go 1
fi

# Step 3:
infof "Update NEWS.md"
countdown "-> Showing git diff NEWS.md" 2
git diff --color=always NEWS.md
if confirm "Prepare NEWS.md with new section?"; then
    release_notes_header="## Version ${release_version} (${release_date})"
    release_commit_msgs="$(git log --format="%s" ${previous_tag}.. | sed -r 's/^/- /')"
    escaped_commit_msgs="$(printf "%q" "${release_commit_msgs}")"
    sed "4i${release_notes_header}\\n\\n${escaped_commit_msgs}\\n\\n"
    countdown "-> Showing git diff NEWS.md" 2
    git diff --color=always NEWS.md
fi
launch_editor NEWS.md 3

# Ensure project builds fine.
infof "Build repoctl"
go build ./... || exit 1
go build . || exit 1
echo "OK."

infof "Test repoctl"
go test ./... || exit 1
go test . || exit 1
echo "OK."

# Step 4:
infof "Create new commit"
git add version.go NEWS.md
countdown "-> Showing git diff --cached" 2
git diff --color=always --cached
confirm "Commit these changes?" || exit 2
git commit -m "Release version ${release_version}"

# Step 5:
infof "Create release archive"
archive_file="repoctl-${release_version}.tar.gz"
git archive --prefix="repoctl-$release_version/" -o "$package_dir/${archive_file}" HEAD
archive_md5sum=$(md5sum "$package_dir/$archive_file" | cut -d' ' -f1)

# Step 6:
infof "Patch PKGBUILD ($package_dir)"
sed -r \
    -e "s/^pkgrel=.*$/pkgrel=1/" \
    -e "s/^pkgver=.*$/pkgver=${release_version}/" \
    -e "s/^md5sums=.*$/md5sums=('${archive_md5sum}')/" \
    -i $package_dir/PKGBUILD || exit 1

infof "Generate .SRCINFO ($package_dir)"
(
    cd $package_dir || exit 1
    makepkg --printsrcinfo > .SRCINFO || exit 1
) || exit 1

infof "Create package ($package_dir)"
(
    cd $package_dir || exit 1
    makepkg || exit 1
) || exit 1

infof "Test package ($package_dir)"
(
    cd $package_dir || exit 1
    sudo pacman -U repoctl-${release_version}-1-x86_64.pkg.tar.zst || exit 1
    expected="repoctl version ${release_version} (${release_date})"
    received="$(/usr/bin/repoctl version | head -1)"
    if [[ "${received}" != "${expected}" ]]; then
        warnf "Unexpected repoctl version output!"
        echo "-- Expected: ${expected}"
        echo "-- Received: ${received}"
        exit 1
    fi
) || exit 1

infof "Commit changes ($package_dir)"
(
    cd $package_dir || exit 1
    git add PKGBUILD .SRCINFO || exit 1
    countdown "-> Showing git diff --cached" 2
    git diff --color=always --cached
    confirm "Commit these changes?" || exit 2
    git commit -m "Update repoctl to version ${release_version}" || exit 1
) || exit 1

# Step 7:
infof "Create new tag"
git tag -a v${release_version} -m "repoctl version ${release_version} release"
echo "Unpushed commits:\n"
git log --oneline --color=always origin/$main_branch..$main_branch | sed -r 's/^/\t/'
echo
confirm "Push $main_branch branch?" || exit 2
git push
confirm "Push v${release_version} tag?" || exit 2
git push origin v${release_version}

# Step 8:
infof "Create new Github release"
echo "Remember to:"
echo " - include release notes from NEWS.md"
echo " - upload the archive from $package_dir"
echo
countdown "-> Launching browser" 1
xdg-open "$github_url/releases/new"
confirm "Is the release ready?"

# Step 9:
infof "Test AUR build again"
(
    set -e
    cd $package_dir
    git clean -xdf
    makepkg

    infof "Push AUR PKGBUILD"
    git push
) || exit 1

infof "Release completed"
