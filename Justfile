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

[doc("versions")]
info:
    {{ if semver_matches(VERSION, ">=42.0.0") == "false" \
        { error("{{YELLOW}}WARNING: version is not semantic{{NORMAL}}") } else { '' } }}
    @printf "{{BOLD}}%s %s-%s{{YELLOW}}%s{{NORMAL}}\n" {{NAME}} {{VERSION}} {{HASH}} \
        {{ if DEBUG == "ON" { "' (debug)'" } else { "" } }}
    @printf "{{ITALIC}}%s %s-%s{{NORMAL}}\n" {{ALT_NAME}} {{ALT_VERSION}} {{ALT_HASH}}

[private]
git:
    {{ if `git rev-parse --symbolic-full-name HEAD` == "HEAD" \
        { error("{{RED}}ERROR: detached HEAD{{NORMAL}}") } else { '' } }}
    @git diff --quiet --exit-code \
        || { echo "{{RED}}ERROR: {{YELLOW}}dirty repository{{NORMAL}}" ; exit 1 ; }

[private]
upstream: git
    @git remote get-url upstream 2>/dev/null \
        || git remote add upstream https://github.com/zyedidia/micro
    @git fetch --tags upstream \
        || { echo "{{RED}}ERROR: {{YELLOW}}fetch{{NORMAL}}" ; exit 1 ; }
    @git tag -d `git tag -l | grep '[a-z]' | grep -v '^v'`
    @git tag -d `git tag -l | grep 'rc'`
    @git maintenance run --auto

[doc("git reset upstream master")]
update: upstream
    @git checkout master \
        || { echo "{{RED}}ERROR: {{YELLOW}}checkout{{NORMAL}}" ; exit 1 ; }
    @git reset --hard upstream/master \
        || { echo "{{RED}}ERROR: {{YELLOW}}reset --hard{{NORMAL}}" ; exit 1 ; }
    @git clean -fd
    @git reset `git tag -l | grep '^v' | grep -v 'rc' | sort -V | tail -1` \
        || { echo "{{RED}}ERROR: {{YELLOW}}reset{{NORMAL}}" ; exit 1 ; }
    @git clean -fd
    @git push origin --tags --force master \
        || { echo "{{RED}}ERROR: {{YELLOW}}push{{NORMAL}}" ; exit 1 ; }
    @git reset --hard
    @git checkout main

[doc("git rebase master")]
upgrade: update
    @git rebase master \
        || { echo "{{RED}}ERROR: {{YELLOW}}rebase{{NORMAL}}" ; exit 1 ; }
    @git push origin main \
        || { echo "{{RED}}ERROR: {{YELLOW}}push{{NORMAL}}" ; exit 1 ; }

[doc("go generate")]
generate:
    @test -d {{CMD}} \
        || { echo "{{RED}}ERROR: {{YELLOW}}invalid input directory{{NORMAL}}" ; exit 1 ; }
    {{GO}} generate {{CMD}}

[doc("go build")]
build: generate
    @.ci/build.sh \
        || { echo "{{RED}}FAILURE: {{YELLOW}}build{{NORMAL}}" ; exit 1 ; }
    @printf {{GREEN}} ; file ./{{NAME}} ; printf {{NORMAL}}

[doc("goos goarch builds")]
release: clean
    @.ci/release.sh \
        || { echo "{{RED}}FAILRE: {{YELLOW}}release{{NORMAL}}" ; exit 1 ; }
    @printf {{GREEN}} ; file ./{{NAME}}-* ; printf {{NORMAL}}

[doc("go install")]
install: build
    GOBIN={{GOBIN}} {{GO}} install -ldflags "{{FLAGS}}" -v {{CMD}}
    @which {{NAME}} >/dev/null 2>/dev/null \
        || { echo "{{RED}}ERROR: {{YELLOW}}executable not found{{NORMAL}}" ; exit 1 ; }
    @printf {{GREEN}} ; which {{NAME}} ; printf {{NORMAL}}

[doc("rm -f …")]
clean:
    @rm -vf {{NAME}}-*
    @rm -vf {{NAME}}
