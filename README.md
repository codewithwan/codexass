# codexass

Small but serious. A Go-based Codex OAuth bridge for local tools and personal AI projects. ⚡

Instead of forcing every project to use API keys, OpenRouter, or extra billing setup, `codexass` lets you use a Codex / ChatGPT OAuth session as the base layer for chat, auth, and usage checks.

## Why this exists

Most people just want this flow:

- log in once
- reuse that session
- connect their own project to an LLM backend
- build their own memory, persona, RAG, or workflow on top

That is what this repo is for.

## What it does

- OAuth login with PKCE
- local session storage
- token refresh
- usage/quota check
- streaming terminal chat
- known model listing
- clean Go layering

## What it does not do

This repo is not trying to be:

- a hosted API
- a full agent framework
- a memory system
- a persona engine
- an official OpenAI SDK replacement

You build those things on top of this base.

## Requirements

- Go `1.25.3`
- a Codex / ChatGPT OAuth-capable account

## Project structure

```text
cmd/
  codexass/

internal/
  cli/
  common/
  config/
  domain/
  infra/
  service/
  ui/

build/
```

## Build 🔧

Build output should go into `build/`, not the repo root.

### Default build command

Windows:

```powershell
./scripts/build.ps1
```

macOS / Linux:

```bash
./scripts/build.sh
```

### Windows

```bash
go build -o ./build/codexass.exe ./cmd/codexass
```

### macOS / Linux

```bash
go build -o ./build/codexass ./cmd/codexass
```

### Run without building

```bash
go run ./cmd/codexass
```

## Commands

### login 🔐

Manual login is the default.

```bash
go run ./cmd/codexass login
```

What happens:

- prints `Login ID`
- prints `Auth URL`
- lets you type `c` + Enter to copy the URL
- does **not** auto-open browser unless you ask for it

Optional:

```bash
go run ./cmd/codexass login --open-browser
go run ./cmd/codexass login --alias my-session
go run ./cmd/codexass login --timeout 300
```

### complete

```bash
go run ./cmd/codexass complete --login-id <LOGIN_ID> --callback-url "<FULL_CALLBACK_URL>"
```

### list

```bash
go run ./cmd/codexass list
```

### usage 📊

```bash
go run ./cmd/codexass usage
go run ./cmd/codexass usage --json
go run ./cmd/codexass usage --session <alias-or-email-or-id>
```

### models 🧠

```bash
go run ./cmd/codexass models
go run ./cmd/codexass models --json
```

### chat 💬

```bash
go run ./cmd/codexass chat
```

Optional:

```bash
go run ./cmd/codexass chat --model gpt-5.3-codex
go run ./cmd/codexass chat --session <alias-or-email-or-id>
```

Chat commands:

- `/exit`
- `/quit`
- `/reset`

## Session storage

Default session location:

- Windows: `%APPDATA%\\codexass`
- Unix-like: `~/.codexass`

Override it with:

```bash
go run ./cmd/codexass --store-dir "/custom/path" list
```

Stored files:

- `auth.json`
- `pending-oauth/*.json`

## Models

`codexass models` currently shows a known supported list.

Right now it is not doing live account-aware model discovery, so:

- some models may vary by account
- `gpt-5.3-codex` is the best default for now

## Limits

- backend behavior may change over time
- this should be treated as a base repo, not a guaranteed public API
- account access and model availability can vary

## Build output policy

Please keep generated binaries inside `build/`.

Do not leave:

- `.exe` files in the repo root
- random binaries mixed into source folders

Keep the repo clean. 🧼

## Next ideas

- live model discovery
- local HTTP bridge
- config file support
- richer terminal UI

If you want memory, personas, custom assistants, or app-specific workflow logic, build them above `codexass`, not inside the core.
