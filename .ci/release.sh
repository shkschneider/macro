#!/usr/bin/env bash
set -eu

command -v git >/dev/null || exit 1
#git diff --quiet --exit-code || { echo "Dirty repository!" >&2 ; exit 1 ; }
command -v go >/dev/null || exit 1

BIN="macro"
[ -d "./cmd/${BIN}" ] || exit 2
echo ":: ${BIN}='./cmd/${BIN}'"
VERSION="${VERSION:-$(git tag -l '[^v]*' | sort -V | tail -1)}"
echo ":: VERSION='${VERSION}'"
HASH="${HASH:-$(git rev-parse --short HEAD)}"
echo ":: HASH='${HASH}'"
DATE="${DATE:-$(date +%F)}"
echo ":: DATE='${DATE}'"
DEBUG="${GODEBUG:-false}"
echo ":: DEBUG='${DEBUG}'"
ARGS="${@:--v}"
echo ":: ARGS='${ARGS}'"
FLAGS="\
 -s -w -extldflags=-static \
 -X 'github.com/zyedidia/micro/v2/internal/util.Version=${VERSION}' \
 -X 'github.com/zyedidia/micro/v2/internal/util.CommitHash=${HASH}' \
 -X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=${DATE}' \
 -X 'github.com/zyedidia/micro/v2/internal/util.Debug=${DEBUG}' \
"

for target in {linux,darwin,windows}/{amd64,arm64} ; do
    goos=$(dirname "$target")
    goarch=$(basename "$target")
    printf "<= %s\n" "$goos/$goarch go build ... ./cmd/${BIN}"
    GOOS=$goos GOARCH=$goarch go build \
        -o ./${BIN}-$goos-$goarch \
        -trimpath \
        -ldflags "${FLAGS}" \
        ${ARGS} ./cmd/${BIN}
    printf "=> %s\n" "./${BIN}-$goos-$goarch"
done

command -v gh >/dev/null || exit
gh release create "${VERSION}" \
    --latest=true \
    --generate-notes --fail-on-no-commits \
    ./macro-*

exit 0

# EOF
