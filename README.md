# Macro

> A modern and intuitive terminal-based text editor.

**Macro** is a fork of [Micro editor](https://github.com/zyedidia/micro) by _Zachary Yedidia_.

Micro is a fantastic minimalist text editor for the terminal,
written in [Go](https://go.dev) and relying on the [Tcell](https://github.com/gdamore/tcell) library.
But it lacks activity (hundreds of issues and pull requests open).

Read more at [Micro's README](https://github.com/zyedidia/micro/blob/master/README.md).

## Fork

Being [Open-Source](https://opensource.org) (MIT License),
Micro allows me to learn Go and how an editor works,
And shape it to my liking.

So far I worked on:
- now able to open a directory (prompts for file with a fuzzy selector)
- explorer to switch file
- new commands
- …

Version of Micro "v2.0.0+" was prefixed with a magic "4" to get Macro "v42.0.0"+.

## Try

```
git clone https://github.com/shkschneider/macro
cd macro
just build
./macro
```

## Upstream

- _branch:main_ HEAD
- _branch:master_ upstream
- Micro -> [Macro](https://github.com/shkschneider/macro)
- Makefile -> [Justfile](https://github.com/casey/just)
- ./cmd/micro -> ./cmd/macro
