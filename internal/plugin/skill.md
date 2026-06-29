---
name: neetodesk
description: >
  Manage NeetoDesk from the command line.
  Use when the user asks about operations exposed by the NeetoDesk CLI.
---

## Prerequisites

Run `neetodesk doctor` to check authentication and connectivity.
If not authenticated, run `neetodesk login`.

## Authentication & multi-subdomain

Credentials for every logged-in subdomain are stored together in
`~/.config/neetodesk/auth.json`. A command that talks to the API picks which
subdomain to use by these rules:

- 0 subdomains authenticated → every credential-using command errors with
  "Not authenticated. Run 'neetodesk login' to authenticate.".
- 1 subdomain authenticated → that one is the implicit default; `--subdomain`
  may be omitted.
- 2+ subdomains authenticated → **`--subdomain <name>` is required** on every
  credential-using command, including `doctor`. The error lists every
  authenticated subdomain so the agent can offer a choice.

`login` / `logout` / `whoami` have dedicated behavior:

| Command | Behavior |
|---|---|
| `neetodesk login --subdomain <name>` | Adds or refreshes the entry for `<name>`. No flag → prompts for the subdomain. |
| `neetodesk logout --subdomain <name>` | Removes that one entry. |
| `neetodesk logout --all` | Removes every entry. |
| `neetodesk logout` (no flag) | Removes the only entry if exactly one is logged in; errors if multiple. |
| `neetodesk whoami` | Lists every logged-in account. Marks the entry `(default)` when exactly one. |
| `neetodesk whoami --subdomain <name>` | Shows just that one. |

## Global flags (persistent on every command)

| Flag | Purpose |
|---|---|
| `--subdomain <name>` | Select which logged-in subdomain the command targets. Required when multiple are logged in. |
| `--json` | Force JSON envelope output even on a TTY. |
| `--quiet` | Emit only the raw payload — no envelope, no breadcrumbs. For action commands (create/update), emits just the resource identifier; `delete` emits `success`. Designed for scripting. |
| `--toon` | Emit TOON (Token Optimized Output Notation). Preferred for feeding list/show output back to an LLM; ~30–60% fewer tokens than JSON. |

Precedence if multiple are set: `--toon` > `--quiet` > `--json` > pretty.

## Output modes & response envelope

**Pretty (default on a TTY)** — tables for arrays, key-value for objects,
breadcrumbs appended. Not intended for machine consumption.

**JSON envelope** (non-TTY, or `--json`):
```json
{
  "data": <resource body>,
  "breadcrumbs": [{ "label": "List", "command": "neetodesk <resource> list" }],
  "pagination": {
    "current_page_number": 1,
    "total_pages": 10,
    "total_records": 250
  }
}
```
`breadcrumbs` is omitted when empty. `pagination` is present only for list
commands.

**Quiet** (`--quiet`) — `data` contents only, no envelope. For action
commands `PrintQuiet` unwraps a single-key wrapper and prints the first of
`sid` / `id` / `name`. For `delete` it prints `success`.

**TOON** (`--toon`) — same data as JSON, re-encoded into TOON. Shape is
equivalent but whitespace/keys are compressed. Parse by re-reading keys as
you would JSON.

### Pagination

List commands accept `--page` (1-indexed) and `--page-size` (max 100).
The envelope's `pagination` field always exposes:
`current_page_number`, `total_pages`, `total_records`. Agents should loop
by incrementing `--page` until `current_page_number == total_pages`.

## Discovery

The full, always-accurate command tree (including any flags added after
this skill was built) is available as JSON:

```bash
neetodesk commands
```

Each catalog entry has `command`, `description`, optional `flags` (with
`name`, `type`, `default`, `description`, `required`), and `subcommands`.
Use this whenever a user asks about a flag or command not covered below.

## Diagnostics & IDE setup

| Command | Purpose |
|---|---|
| `doctor` | Auth check + API reachability + version. Uses `--subdomain` when multiple are logged in. |
| `version` | Print CLI version / commit / build date. |
| `commands` | Emit the full command/flag catalog as JSON. |
| `setup claude` | Install NeetoDesk plugin into Claude Code (`plugin.json`, hooks, this SKILL.md). |
| `setup cursor` / `windsurf` / `copilot` / `gemini` / `codex` | Write IDE-specific NeetoDesk rule files. |

## Environment variable override

Set `NEETODESK_BASE_URL` to point the CLI at a staging or local server:

```bash
export NEETODESK_BASE_URL=http://acme.lvh.me:8980
neetodesk login --subdomain acme
```

## Error surface

Every command exits non-zero on failure and writes a single-line message to
stderr. Common errors the agent should expect:

- `Not authenticated. Run 'neetodesk login' to authenticate.` — empty credential store.
- `Multiple subdomains authenticated (acme, beta); specify --subdomain.` — pick one.
- `Not authenticated for "foo". Authenticated subdomains: acme, beta.` — bad `--subdomain`.
- `required flag(s) "xxx" not set` (from cobra) — missing required flag.
- API errors come through with the server's message body; inspect the
  JSON envelope (or the `--quiet` payload) for `error` / `errors` / `notice`
  keys and any suggestions the API returns.

## Product-specific commands

This skeleton CLI does not yet ship product-specific resource commands.
Run `neetodesk commands` to see what is currently available, and
refer to the CLI's own docs for the full command reference once it grows.
