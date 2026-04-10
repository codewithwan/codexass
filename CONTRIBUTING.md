# Contributing

Thanks for contributing.

## Rules

- keep it clean
- keep it small
- keep behavior predictable
- do not add random framework-style abstractions unless they solve a real problem

## Project direction

`codexass` is a base repo for:

- OAuth auth
- session storage
- chat transport
- usage checks
- model listing

It is **not** the place for:

- memory engines
- persona systems
- agent workflow logic
- app-specific product behavior

Build those on top.

## Code expectations

- no source file above the agreed file-size limit
- no sloppy generated binaries in the repo root
- keep build artifacts in `build/`
- keep config in `internal/config`
- prefer typed structs over loose dynamic parsing
- preserve existing CLI behavior unless the change is intentional

## Before opening a change

Please make sure:

1. code is formatted
2. project builds successfully
3. main commands still work
4. docs are updated if behavior changes

## Useful checks

```bash
go build ./cmd/codexass
go run ./cmd/codexass list
go run ./cmd/codexass models
go run ./cmd/codexass chat
```

## Build output

Preferred build commands:

### Windows

```powershell
./scripts/build.ps1
```

### macOS / Linux

```bash
./scripts/build.sh
```

Manual examples:

```bash
go build -o ./build/codexass.exe ./cmd/codexass
```

```bash
go build -o ./build/codexass ./cmd/codexass
```

## Pull request style

- explain the problem briefly
- explain the fix briefly
- keep PRs focused
- do not mix cleanup, refactor, and feature work unless they truly belong together

That is enough. Clean PRs win.
