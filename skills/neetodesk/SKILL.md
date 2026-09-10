---
name: neetodesk
description: >
  Manage NeetoDesk helpdesk data from the command line: tickets, ticket comments and
  drafts, customers, contact forms, team members (agents), and agent/group/ticket/CSAT
  reports. Use when the user asks about support tickets, replying to or leaving an
  internal note on a ticket, creating a customer or inviting a team member, checking
  which contact forms are enabled, or pulling agent, group, survey, or ticket-volume
  reports for NeetoDesk.
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
| `--toon` | Emit TOON (Token Optimized Output Notation). Preferred for feeding list/show/report output back to an LLM; ~30–60% fewer tokens than JSON. |

Precedence if multiple are set: `--toon` > `--quiet` > `--json` > pretty.

## Output modes & response envelope

**Pretty (default on a TTY)** — tables for arrays, key-value for objects,
breadcrumbs appended. Good for a human reading the terminal directly; not
intended for machine consumption.

**JSON envelope** (non-TTY, or `--json`):
```json
{
  "data": <resource body>,
  "breadcrumbs": [{ "label": "List", "command": "neetodesk tickets list" }],
  "pagination": {
    "current_page_number": 1,
    "total_pages": 10,
    "total_records": 250
  }
}
```
`breadcrumbs` is omitted when empty. `pagination` is present only for
paginated list commands. Use `--json` when the caller needs the pagination
metadata alongside the data (e.g. deciding whether to fetch another page).

**Quiet** (`--quiet`) — `data` contents only, no envelope. For action
commands (create/update) it prints just the resource identifier. For
`delete` it prints `success`. Use this to pipe an id into another command
or into `jq`.

**TOON** (`--toon`) — same data as JSON, re-encoded into TOON. Shape is
equivalent but whitespace/keys are compressed. This is the best default
when an agent is going to read a list, a ticket, or a report back into its
own context — it carries the same information as JSON for noticeably fewer
tokens.

### Pagination

List commands that paginate accept `--page` (1-indexed) and `--page-size`
(max 100): `tickets list`, `tickets comments list`, `team-members list`,
`reports agents`, and `reports groups`. The envelope's `pagination` field
always exposes `current_page_number`, `total_pages`, `total_records`.
Agents should loop by incrementing `--page` until
`current_page_number == total_pages`. `forms list` and the other report
commands (`surveys`, `tickets`, `ticket-time-series`) return their full
result in one call and do not paginate.

## Discovery

The full, always-accurate command tree (including any flags added after
this skill was built) is available as JSON:

```bash
neetodesk commands
```

Each catalog entry has `command`, `description`, optional `flags` (with
`name`, `type`, `default`, `description`, `required`), and `subcommands`.
Use this whenever a user asks about a flag or command not covered below.

## Command reference

Required flags are marked with `*`.

### Tickets

A ticket is the central helpdesk work item: a customer's request, tracked
through a status, priority, category, and an assigned agent/group. The
description given at creation is stored as the ticket's first comment.

| Command | Positional | Flags |
|---|---|---|
| `tickets list` | — | `--status` (comma-separated, e.g. `open,pending`), `--page`, `--page-size` |
| `tickets show` | `<id>` | — |
| `tickets create` | — | `--email*` (customer email), `--subject*`, `--description*`, `--name` (customer name), `--status`, `--priority` (`low`/`medium`/`high`/`urgent`), `--category`, `--agent-id`, `--group-id`, `--assignee-email` |
| `tickets update` | `<id>` | `--subject`, `--description`, `--status`, `--priority`, `--category`, `--agent-id`, `--group-id`, `--assignee-email` (all partial — only the flags set are sent) |

An agent can be assigned either by `--agent-id` (the team member's id) or
by `--assignee-email`; same for a group via `--group-id`.

### Ticket comments

Comments are the replies and internal notes on a ticket. A `reply` is
visible to the customer; a `note` is internal to the workspace only.

| Command | Positional | Flags |
|---|---|---|
| `tickets comments list` | `<ticket-id>` | `--page`, `--page-size` |
| `tickets comments show` | `<ticket-id> <comment-id>` | — |
| `tickets comments create` | `<ticket-id>` | `--content*` (HTML), `--comment-type` (`reply` or `note`, default `reply`) |

### Ticket drafts

A draft is an unsent reply or note. A ticket holds at most one draft per
comment type — creating a new one for the same type replaces the previous
draft. There is no command to read a draft back; it surfaces to the agent
in the NeetoDesk UI.

| Command | Positional | Flags |
|---|---|---|
| `tickets drafts create` | `<ticket-id>` | `--content*` (HTML), `--comment-type` (`reply` or `note`, default `reply`), `--author-email` (agent the draft is written as) |

### Customers

Customers are the people who raise tickets. Only creation is currently
supported — there is no `list`, `show`, `update`, or `delete` for
customers yet; look up a customer's tickets by their email through
`tickets list` / `tickets show` instead.

| Command | Positional | Flags |
|---|---|---|
| `customers create` | — | `--email*`, `--first-name`, `--last-name`, `--phone`, `--language` (preferred language), `--time-zone` (e.g. `America/New_York`), `--description`, `--company-id` |

### Team members

Team members are the workspace's agents and admins.

`team-members list` returns newest first by default, matching the neeto-desk
web UI. Pass `--sort email` or `--sort last_name` to list them alphabetically,
or `--order asc` for oldest first.

| Command | Positional | Flags |
|---|---|---|
| `team-members list` | — | `--email` (filter), `--sort` (`created_at`, `updated_at`, `email`, `first_name` or `last_name`, default `created_at`), `--order` (`asc` or `desc`, default `desc`), `--page`, `--page-size` |
| `team-members show` | `<id>` | — |
| `team-members create` | — | `--email*` (repeat the flag for multiple invitees), `--role*` (e.g. `agent`, `admin`), `--send-invitation-email` (bool, default `true`) |
| `team-members update` | `<id>` | `--email`, `--first-name`, `--last-name`, `--role`, `--time-zone` (all partial) |
| `team-members delete` | `<id>` | — |

### Forms

| Command | Positional | Flags |
|---|---|---|
| `forms list` | — | — |

Lists the workspace's currently enabled contact forms — the web forms
customers use to raise a ticket. Disabled forms are omitted.

### Reports

Reports aggregate ticket, agent, group, and survey metrics over a date
range. `--range-type` accepts values like `last_7_days` or `last_30_days`;
pass `custom` together with `--start-date` and `--end-date` (`YYYY-MM-DD`)
for an exact window. Omitting all three flags uses NeetoDesk's own default
range.

| Command | Positional | Flags | What it returns |
|---|---|---|---|
| `reports agents` | — | `--range-type`, `--start-date`, `--end-date`, `--page`, `--page-size` | One row per agent: tickets handled, response/resolution times, and similar performance metrics for the period. |
| `reports groups` | — | `--range-type`, `--start-date`, `--end-date`, `--page`, `--page-size` | The same performance metrics rolled up per group/team instead of per agent. |
| `reports surveys` | — | `--range-type`, `--start-date`, `--end-date` | Customer satisfaction (CSAT) survey results for the period. |
| `reports tickets` | — | `--range-type`, `--start-date`, `--end-date` | Ticket status counts for the period versus the prior period, with the percentage change per status. |
| `reports ticket-time-series` | — | `--range-type`, `--start-date`, `--end-date` | Day-by-day ticket creation and closure counts across the period, suitable for charting a trend. |

## Diagnostics, shell setup & IDE setup

| Command | Purpose |
|---|---|
| `doctor` | Auth check + API reachability + version. Uses `--subdomain` when multiple are logged in. |
| `version` | Print CLI version / commit / build date. |
| `commands` | Emit the full command/flag catalog as JSON. |
| `completion bash` / `zsh` / `fish` / `powershell` | Install shell tab-completion for `neetodesk`. Pass `--print` to print the script instead of installing it. |
| `setup claude` | Install the NeetoDesk plugin into Claude Code (`plugin.json`, hooks, this SKILL.md). |
| `setup cursor` / `windsurf` / `copilot` / `gemini` / `codex` | Write NeetoDesk rule files into the current project directory; re-run after an upgrade to refresh them. |
| `update` | Update the CLI to the latest version (auto-detects brew / shell / PowerShell install). |

## Environment variable override

Set `NEETODESK_BASE_URL` to point the CLI at a staging or local server:

```bash
export NEETODESK_BASE_URL=http://acme.lvh.me:8980
neetodesk login --subdomain acme
```

## Common workflows

### Authenticate into a subdomain and confirm access
```bash
neetodesk login --subdomain acme
neetodesk doctor
neetodesk whoami
```

### Create a ticket, then reply to it
```bash
neetodesk tickets create \
  --email customer@example.com --name "Jane Doe" \
  --subject "Cannot reset password" \
  --description "Reset link in the email 404s." \
  --priority high --quiet
# → prints the new ticket id

neetodesk tickets comments create <ticket-id> \
  --content "<p>Thanks for reporting this — looking into it now.</p>" \
  --comment-type reply
```

### Leave an internal note without notifying the customer
```bash
neetodesk tickets comments create <ticket-id> \
  --content "Escalated to engineering, see INFRA-482." \
  --comment-type note
```

### Draft a reply before sending it
```bash
neetodesk tickets drafts create <ticket-id> \
  --content "<p>Draft: refund has been processed.</p>" \
  --author-email agent@acme.com
```

### Triage open tickets and page through the list
```bash
neetodesk tickets list --status open,pending --page-size 50 --toon
# repeat with --page 2, --page 3, ... until current_page_number == total_pages
```

### Reassign and re-prioritize a ticket
```bash
neetodesk tickets update <ticket-id> \
  --assignee-email agent@acme.com --priority urgent --status open
```

### Invite a new agent
```bash
neetodesk team-members create \
  --email new.agent@acme.com --role agent
```

### Pull last month's agent performance report
```bash
neetodesk reports agents --range-type last_30_days --toon
```

### Check ticket volume trends for a custom window
```bash
neetodesk reports ticket-time-series \
  --range-type custom --start-date 2026-08-01 --end-date 2026-08-31 --toon
```

### See which contact forms are live before pointing a customer at one
```bash
neetodesk forms list
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

## Conventions

- Dates: `YYYY-MM-DD`. Time zones: IANA names (`America/New_York`).
- Comment/draft content is HTML, not plain text or Markdown.
- IDs: tickets, comments, team members and customers are addressed by the
  `id` field returned in their `show`/`list`/`create` response.
- For any flag or field not covered above, `neetodesk commands` is
  authoritative.
