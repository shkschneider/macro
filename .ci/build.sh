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

printf "<= %s\n" "go build -o ./${BIN} -trimpath -ldflags '...' ${ARGS} ./cmd/${BIN}"

go build -o "./${BIN}" \
 -trimpath \
 -ldflags "\
 -s -w -extldflags=-static \
 -X 'github.com/zyedidia/micro/v2/internal/util.Version=${VERSION}' \
 -X 'github.com/zyedidia/micro/v2/internal/util.CommitHash=${HASH}' \
 -X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=${DATE}' \
 -X 'github.com/zyedidia/micro/v2/internal/util.Debug=${DEBUG}' \
 " "${ARGS}" "./cmd/${BIN}" || exit 3

printf "=> %s\n" "$(file ./${BIN})"
exit 0

# EOF
