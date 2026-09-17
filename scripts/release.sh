#!/usr/bin/env bash
# release.sh: cut a Ginkgo release.  The Release workflow (.github/workflows/release.yml) runs this
# after the test workflow passes; see RELEASING.md.
#
# Usage: scripts/release.sh patch|minor
#
# One run: release ## Unreleased as vX.Y.Z, commit and tag it, push both, and create the GitHub
# release.  Re-running after a failure resumes: once the tag is on origin, the run builds from the
# tag and skips whatever is already released.
#
# GINKGO_RELEASE_DRY_RUN=1 does everything locally, then stops short of pushing and creating the
# GitHub release.  A development aid only.
set -euo pipefail

bump=${1:-}
[[ "$bump" == patch || "$bump" == minor ]] || { echo "usage: $0 patch|minor" >&2; exit 2; }
dry_run=${GINKGO_RELEASE_DRY_RUN:-}

cd "$(dirname "${BASH_SOURCE[0]}")/.."

say() { printf '\n==> %s\n' "$*"; }
fail() { printf 'release failed: %s\n' "$*" >&2; exit 1; }
tool() { .release/release-tool "$@"; }

[[ "$(git rev-parse --abbrev-ref HEAD)" == master ]] || fail "releases are cut from master"
[[ -z "$(git status --porcelain)" ]] || fail "the working tree is not clean"

mkdir -p .release && go build -o .release/release-tool ./scripts/release
version=$(tool next "$bump")
tag="v$version"

# ls-remote exits 2 when the tag is absent; anything else nonzero means origin could not be asked.
tag_status=0
git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null || tag_status=$?
[[ $tag_status == 0 || $tag_status == 2 ]] || fail "could not check origin for $tag"

if [[ $tag_status == 0 ]]; then
	say "$tag is already on origin - resuming the release from it"
	git fetch --force origin "refs/tags/$tag:refs/tags/$tag"
	git checkout --quiet --detach "$tag"
	[[ "$(tool current)" == "$version" ]] || fail "$tag does not have VERSION $version"
else
	say "Preparing $tag"
	tool prepare "$version"
	go build ./... # the version file is Go source; never ship a release that does not compile
	git -c user.name="github-actions[bot]" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
		commit --quiet --all --message "$tag"
	git tag "$tag"
	git show --stat HEAD
	if [[ -n "$dry_run" ]]; then
		say "dry run: not pushing master and $tag"
	else
		# Atomic: origin gets the release commit and its tag together or not at all, so a tag on
		# origin is proof the commit is there too.  Fast-forward only - fails if master moved.
		git push --atomic origin HEAD:refs/heads/master "refs/tags/$tag"
	fi
fi

tool notes "$version" >.release/notes.md
say "Release notes for $tag"
cat .release/notes.md

if [[ -n "$dry_run" ]]; then
	say "dry run: not creating the GitHub release $tag"
elif gh release view "$tag" >/dev/null 2>&1; then
	say "The GitHub release $tag already exists - refreshing its notes"
	gh release edit "$tag" --notes-file .release/notes.md
else
	say "Creating the GitHub release $tag"
	gh release create "$tag" --verify-tag --title "$tag" --notes-file .release/notes.md
fi

say "Released $tag${dry_run:+ (dry run)}"
