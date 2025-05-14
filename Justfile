UPSTREAM    := "https://github.com/zyedidia/micro"
ALT_NAME    := "micro"
ALT_VERSION := shell("git tag -l 'v*' | sort -V | tail -1 | cut -c2-")
ALT_HASH    := shell("git show-ref --abbrev 'v" + ALT_VERSION + "' | cut -d' ' -f1")

NAME    := replace(ALT_NAME, "i", "a") # macro
GO      := require("go")
VERSION := shell("git tag -l '[^v]*' | sort -V | tail -1")
HASH    := shell("git rev-parse --short HEAD")
DATE    := datetime_utc("%F") #ISO8601
DEBUG   := if lowercase(env("GODEBUG", "")) != "" { "ON" } else { "OFF" }
FLAGS   := "-s -w \
-X 'github.com/zyedidia/micro/v2/internal/util.Version=" + VERSION + "' \
-X 'github.com/zyedidia/micro/v2/internal/util.CommitHash=" + HASH + "' \
-X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=" + DATE + "' \
-X 'github.com/zyedidia/micro/v2/internal/util.Debug=" + DEBUG + "'"
GOBIN   := env("GOBIN", executable_directory())
CMD     := "." / "cmd" / NAME

alias i := info
alias u := update
alias b := build
alias c := clean

[private]
default:
    @just --list

[doc("version")]
info:
    {{ if semver_matches(VERSION, ">=42.0.0") == "false" { error("Version is not semantic!") } else { "" } }}
    @printf "{{BOLD}}%s %s-%s{{YELLOW}}%s{{NORMAL}}\n" {{NAME}} {{VERSION}} {{HASH}} \
        {{ if DEBUG == "ON" { "' (debug)'" } else { "" } }}
    @printf "{{ITALIC}}%s %s-%s{{NORMAL}}\n" {{ALT_NAME}} {{ALT_VERSION}} {{ALT_HASH}}

[doc("git pull upstream")]
update:
    @git remote get-url upstream || git remote add upstream "{{UPSTREAM}}"
    git fetch upstream
    git pull upstream master
    git tag -l 'v*' | sort -V | tail -1

[doc("go generate + go build")]
build: clean
    @test -d {{CMD}} || { echo "{{RED}}Invalid input directory!{{NORMAL}}" ; exit 1 ; }
    {{GO}} build -trimpath -ldflags "{{FLAGS}}" -v {{CMD}}
    @printf {{GREEN}} ; file ./{{NAME}} ; printf {{NORMAL}}

[doc("go install")]
install: build
    GOBIN={{GOBIN}} {{GO}} install -ldflags "{{FLAGS}}" -v {{CMD}}
    @which {{NAME}} >/dev/null 2>/dev/null || { echo "{{RED}}Executable not found!{{NORMAL}}" ; exit 1 ; }
    @printf {{GREEN}} ; which {{NAME}} ; printf {{NORMAL}}

[doc("uninstall")]
uninstall:
    #!/usr/bin/env bash
    bin="$(which {{NAME}} 2>/dev/null)"
    test -x "$bin" && rm -vf "$bin" || { echo "{{RED}}Executable not found!{{NORMAL}}" ; exit 1 ; }

[doc("goos goarch builds")]
release: clean
    #!/usr/bin/env bash
    set -eu
    git diff --quiet --exit-code || { echo "{{RED}}Dirty repository!{{NORMAL}}" >&2 ; exit 1 ; }
    #go tool dist list
    for target in {linux,darwin,windows}/{amd64,arm64} ; do
        goos=$(dirname "$target")
        goarch=$(basename "$target")
        GOOS=$goos GOARCH=$goarch go build \
            -o ./{{NAME}}-$goos-$goarch \
            -trimpath \
            -ldflags "{{FLAGS}}" \
            {{CMD}}
    done
    file ./{{NAME}}-*

[doc("rm -f …")]
clean:
    @rm -vf {{NAME}}-*
    @rm -vf {{NAME}}
