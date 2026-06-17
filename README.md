# go_template

A template for my Go projects.

## Features

- Linting and formatting (golangci-lint v2, prettier).
- Reproducible dev environment (devbox), auto-activated with direnv.
- Dagger-based CI: every gate runs through the local `ci` toolchain
  (`dagger call ci <task>`), composing shared toolchains from
  [go.jacobcolvin.com/x][x] (devbox, goreleaser, security, zizmor).
- Conventional-commit enforcement (commitlint + lefthook git hooks).
- Build and release automation (goreleaser binaries plus Dagger-native
  multi-arch images with cosign keyless signing).
- Cobra command-line interface with [fang][fang]; structured logging and
  profiling from [go.jacobcolvin.com/x][x].
- Taskfile with format, lint, test, check, and build targets.

## Install

```bash
task build          # cross-compile snapshot binaries to ./dist via Dagger
go install ./cmd/go_template
```

## Usage

```bash
go_template hello
go_template version
```

## Development

```bash
devbox install      # provision the toolchain (or run `direnv allow`)
task check          # local gate: lint + test
task check:all      # everything CI runs (adds the Dagger-backed gates)
```

[x]: https://github.com/MacroPower/x
[fang]: https://github.com/charmbracelet/fang
