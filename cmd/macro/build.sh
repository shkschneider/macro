#!/usr/bin/env bash

command -v go >/dev/null || exit 1
command -v git >/dev/null || exit 2
[[ $(realpath "${BASH_SOURCE[0]}") == *"/cmd/macro/${0##*/}" ]] || exit 3

flags=()
flags+=("-s") # strip

micro_tag=$(git tag 2>/dev/null | sort -V | tail -1)
macro_build=$(git rev-list --count "${micro_tag}..HEAD" 2>/dev/null)
[[ -n "$macro_build" ]] || macro_build="0"
if [[ -n "$(git status --porcelain)" ]] ; then
    macro_build="${macro_build}-dev"
else
    flags+=("-w") # no-debug
fi
micro_tag=${micro_tag%%-*}
micro_tag=${micro_tag#v}
[[ -n "$micro_tag" ]] || micro_tag="0.0.0"
flags+=("-X "\
"github.com/micro-editor/micro/v2/internal/util.Version"\
"=${micro_tag}-${macro_build}")

macro_hash=$(git rev-parse --short HEAD 2>/dev/null)
flags+=("-X "\
"github.com/micro-editor/micro/v2/internal/util.CommitHash"\
"=${macro_hash}")

macro_date=$(date +'%F' 2>/dev/null)
flags+=("-X "\
"github.com/micro-editor/micro/v2/internal/util.CompileDate"\
"=${macro_date}")

echo go build -trimpath -ldflags "${flags[*]}" -o macro ./cmd/macro

go build -trimpath -ldflags "${flags[*]}" -o macro ./cmd/macro || exit 4

file ./macro || exit 5

./macro -version || exit 6

# EOF
