# NeetoDesk CLI

A command-line interface for NeetoDesk.

<!-- neeto-cli-commons:installation:start -->
## Installation

### macOS / Linux

**Homebrew (recommended on macOS):**

```bash
brew install neetozone/tap/neetodesk
```

**Shell script:**

```bash
curl -fsSL https://neeto-downloads.s3.amazonaws.com/cli/NeetoDesk/latest/install.sh | sh
```

This verifies the download's SHA-256 checksum against the published `SHA256SUMS`,
then installs to `/usr/local/bin` (may prompt for sudo). Set `NEETODESK_INSTALL_DIR`
to a directory you own to install without sudo.

### Windows

**PowerShell:**

```powershell
irm https://neeto-downloads.s3.amazonaws.com/cli/NeetoDesk/latest/install.ps1 | iex
```

**Command Prompt (CMD):**

```cmd
curl -fsSL https://neeto-downloads.s3.amazonaws.com/cli/NeetoDesk/latest/install.cmd -o install.cmd && install.cmd
```

Both verify the download's SHA-256 checksum before installing to
`%LOCALAPPDATA%\Programs\neetodesk` and adding it to your user PATH. Set
`NEETODESK_INSTALL_DIR` to install somewhere else.
<!-- neeto-cli-commons:installation:end -->

<!-- neeto-cli-commons:verify-installation:start -->
### Verify installation

```bash
neetodesk --help
```
<!-- neeto-cli-commons:verify-installation:end -->

<!-- neeto-cli-commons:ai-coding-assistants:start -->
## AI coding assistants

```bash
neetodesk setup claude      # Register plugin with Claude Code
neetodesk setup cursor      # Write .cursor/rules/neetodesk.mdc
neetodesk setup windsurf    # Write .windsurf/rules/neetodesk.md
neetodesk setup copilot     # Add a NeetoDesk section to .github/copilot-instructions.md
neetodesk setup gemini      # Add a NeetoDesk section to GEMINI.md
neetodesk setup codex       # Add a NeetoDesk section to AGENTS.md
```

Every command except `setup claude` writes into the current project directory, so
run these commands from the root of the project the assistant works in. Re-run
them after every upgrade: `setup cursor` and `setup windsurf` overwrite their rule
file, while `setup copilot`, `setup gemini` and `setup codex` keep the existing
content of their file and replace only the NeetoDesk section instead of adding a
duplicate.
<!-- neeto-cli-commons:ai-coding-assistants:end -->

<!-- neeto-cli-commons:prerequisites:start -->
## Prerequisites (development)

- [Go](https://go.dev/dl/) 1.26.1+
- Access to a NeetoDesk organization
<!-- neeto-cli-commons:prerequisites:end -->

## Development

```bash
git clone https://github.com/neetozone/neeto-desk-cli.git
cd neeto-desk-cli
bin/setup
```

This installs Go dependencies, golangci-lint, configures git hooks, and builds the binary.

<!-- neeto-cli-commons:make-targets:start -->
### Make targets

```bash
make build          # Builds ./neetodesk
make test           # Run tests
make lint           # golangci-lint
make fmt            # gofmt -w
make vet            # go vet
make check          # fmt + vet + test
make install        # Installs to /usr/local/bin
make clean          # Remove built binary
```
<!-- neeto-cli-commons:make-targets:end -->

### Pointing to a local or staging server

Set `NEETODESK_BASE_URL` to override the default `https://<subdomain>.neetodesk.com`:

```bash
export NEETODESK_BASE_URL=http://acme.lvh.me:8980
neetodesk login --subdomain acme
```

<!-- neeto-cli-commons:global-flags:start -->
## Global flags

Every command accepts:

| Flag | Description |
|---|---|
| `--subdomain <name>` | Which logged-in subdomain to use (required when multiple are logged in). |
| `--json` | Force JSON envelope output. |
| `--quiet` | Emit raw data only. Action commands print just the identifier; `delete` prints `success`. |
| `--toon` | TOON (Token-Optimized Output Notation) — compact format for LLMs. |
| `--verbose` | Expand every field of a record instead of a table. |
<!-- neeto-cli-commons:global-flags:end -->

## Adding product-specific commands

See [`docs/adding-commands.md`](docs/adding-commands.md) for the step-by-step
workflow for adding new resource commands that use the built-in auth, HTTP
client, and output helpers.

Quick API wrapper reference: [`docs/api-wrapper-reference.md`](docs/api-wrapper-reference.md).

<!-- neeto-cli-commons:release:start -->
## Release

Releases are cut by BigBinary's CI pipeline defined in
`.neetoci/release.yml`. Merging a PR with a `major` / `minor` / `patch`
label to `main` triggers the shared release script published by
`neeto-cli-commons`, which bumps and tags VERSION, runs GoReleaser,
uploads artifacts to `s3://neeto-downloads/cli/NeetoDesk/`, updates the
Homebrew tap (`neetozone/tap`), and pushes the version bump commit
straight to `main`.
<!-- neeto-cli-commons:release:end -->
