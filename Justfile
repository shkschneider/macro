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
CFLAGS  := if env("CGO_ENABLED", "0") == "0" { "-extldflags=-static" } else { "" }
FLAGS   := "-s -w " + CFLAGS + " \
-X 'github.com/zyedidia/micro/v2/internal/util.Version=" + VERSION + "' \
-X 'github.com/zyedidia/micro/v2/internal/util.CommitHash=" + HASH + "' \
-X 'github.com/zyedidia/micro/v2/internal/util.CompileDate=" + DATE + "' \
-X 'github.com/zyedidia/micro/v2/internal/util.Debug=" + DEBUG + "' \
"
GOBIN   := env("GOBIN", executable_directory())
CMD     := "." / "cmd" / NAME

alias i := info
alias b := build
alias c := clean

[private]
default:
    @just --list

[doc("version")]
info:
    {{ if semver_matches(VERSION, ">=42.0.0") == "false" { error("{{RED}}Version is not semantic!{{NORMAL}}") } else { '' } }}
    @printf "{{BOLD}}%s %s-%s{{YELLOW}}%s{{NORMAL}}\n" {{NAME}} {{VERSION}} {{HASH}} \
        {{ if DEBUG == "ON" { "' (debug)'" } else { "" } }}
    @printf "{{ITALIC}}%s %s-%s{{NORMAL}}\n" {{ALT_NAME}} {{ALT_VERSION}} {{ALT_HASH}}

[doc("go generate")]
generate:
    @test -d {{CMD}} || { echo "{{RED}}Invalid input directory!{{NORMAL}}" ; exit 1 ; }
    {{GO}} generate {{CMD}}

[doc("go build")]
build: generate
    @.ci/build.sh || { echo "{{RED}}Build FAILED" ; exit 1 ; }
    @printf {{GREEN}} ; file ./{{NAME}} ; printf {{NORMAL}}

[doc("goos goarch builds")]
release: clean
    @.ci/release.sh || { echo "{{RED}}Release FAILED" ; exit 1 ; }
    @printf {{GREEN}} ; file ./{{NAME}}-* ; printf {{NORMAL}}

[doc("go install")]
install: build
    GOBIN={{GOBIN}} {{GO}} install -ldflags "{{FLAGS}}" -v {{CMD}}
    @which {{NAME}} >/dev/null 2>/dev/null || { echo "{{RED}}Executable not found!{{NORMAL}}" ; exit 1 ; }
    @printf {{GREEN}} ; which {{NAME}} ; printf {{NORMAL}}

[doc("rm -f …")]
clean:
    @rm -vf {{NAME}}-*
    @rm -vf {{NAME}}
