#!/usr/bin/env bash
set -o errexit
set -o nounset

GINKGO=${GINKGO:-ginkgo}

input=${1:-""}
for format in "csv" "json"; do
    set -o xtrace
    output="$(dirname $input)/$(basename $input).$format"
    tmp=$(mktemp ginkgo-outline-test.XXX)
    if "$GINKGO" outline --format="$format" "$input" 1>"$tmp"
    then mv "$tmp" "$output"
    else rm "$tmp"
    set +o xtrace
    fi
done
