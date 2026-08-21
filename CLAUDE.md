# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`eds` (Evolution DevServices CLI, repo: `git@github.com:cloud-ru/evolution-devservices-cli.git`)
is a Go CLI client for cloud.ru "developer tools" — a platform CLI
that currently drives two independent products:

- the **Repo product** (git repositories in the cloud, at `devtools.api.cloud.ru/repo`) — `eds repo *`.
- **Workflow Studio** (wires a repository + branch to a deploy pipeline and
  publishes it) — `eds wf app|run|job *`.

Both products are authenticated with the same API key (`--api-key`/`EDS_API_KEY`),
sent as `X-API-KEY`. `--project`/`EDS_PROJECT_ID` is the only other platform-level
(shared, unprefixed) setting.

The CLI is explicitly designed to be **agent-friendly**: every command has
stable `--json` output, config comes from env vars, and `skill/SKILL.md`
documents the CLI's contract for AI agents driving it. The primary
agent-facing scenario is *ship a vibe-coded app*: `eds repo create` +
`git push` gets code hosted, `eds wf app create` + `eds wf app deploy`
publishes it, and `eds wf app show` / `eds wf run show` /
`eds wf job logs` track the rollout.

## Commands

```bash
make build         # build for current platform -> ./bin/eds
make build-all      # cross-compile darwin/linux x amd64/arm64 -> ./dist/
make test           # go test ./...
make lint           
make tidy           # go mod tidy
make clean          # remove ./bin and ./dist
```

There are currently no `_test.go` files in the repo, so `make test` is a no-op
until tests are added. `make all` runs `lint test build` in sequence.

Building directly without make:

```bash
go build -o bin/eds .
go run . <command>
```

Version is injected at build time via `-ldflags "-X main.version=..."` (see
`Makefile`'s `LDFLAGS`); running via `go run` yields version `dev`. The Go
module path is `github.com/cloud-ru/evolution-devservices-cli`, matching the
repo's actual GitHub location — unlike the earlier `evo` rebrand (which kept
the pre-rename module path as a cosmetic-only, minimal-diff change), this
rename moved the repo itself to a new GitHub org/name, so `go install
github.com/cloud-ru/evolution-devservices-cli@latest` needs the module path
to match. The internal `Config` struct field names and JSON keys still keep
their older spelling (see the `internal/config/config.go` entry below) —
that part of the prior minimal-diff choice still applies, since those are
genuinely internal and not tied to where the repo lives.

## Architecture

Cobra-based CLI, one command per file under `cmd/`, thin API clients under
`internal/repoapi/` (Repo product) and `internal/workflow_client/` (Workflow
Studio).

- `main.go` — entry point, wires `main.version` (ldflags) into `cmd.SetVersion`.
- `cmd/root.go` — builds the root `eds` command, registers global persistent
  flags. Product-scoped: `--repo-api-url`, `--repo-git-host`
  (Repo product); `--wf-api-url` (Workflow Studio). Platform-level (shared,
  unprefixed): `--project`, `--api-key`. Plus `--json`, `--quiet`.
- `cmd/helpers.go` — `resolveContext(cmd)` is the entry point every subcommand
  calls first. It loads config, layers flag overrides on top (flags > env >
  file > defaults), and builds a `runtimeContext{Cfg, API, Printer, Quiet,
  WorkflowClient}`. `requireAPIKey()` guards Repo-product commands.
  `ensureWorkflowAuth(ctx)` guards Workflow Studio commands: it validates
  that `APIKey` is set (the same key is used for both products). Both helpers
  return the same friendly error when the key is missing.
- `cmd/login.go`, `cmd/config.go`, `cmd/version.go`, `cmd/repo.go` — one
  subcommand tree each. `repo.go` holds `eds repo list|create|show|delete|clone`
  plus `resolveRepoID`, which lets users pass either a UUID or a repo name
  (falls back to listing + case-insensitive name match when the arg doesn't
  look like a UUID). `login.go` accepts `--api-key` plus optional product URLs.
- `cmd/wf.go` — the `eds wf` parent command; just groups `newAppCmd()`,
  `newRunCmd()`, `newJobCmd()` as children. Doesn't hold any logic itself.
- `cmd/app.go`, `cmd/run.go`, `cmd/job.go` — Workflow Studio subcommand
  trees, nested under `eds wf`. `app.go` holds `eds wf app create|list|show|
  update|delete|deploy|deployments` (an **application** wires a
  repo+branch to a deploy pipeline; a **deployment** runs that pipeline and
  publishes it). `app create --repository` reuses `resolveRepoID` from
  `repo.go` to accept either a Repo-product UUID or name. `run.go`/`job.go`
  are the lower-level primitives for inspecting runs and streaming job
  logs. `job.go`'s `logs` command reads the job-logs endpoint as SSE and
  prints each `data:` line to stdout.
  Note (observed empirically, not documented in the spec): `application.status`
  and `application.run.status` are separate lifecycles. Creating an
  application auto-triggers its first run; even after `run.status` reaches
  `done`, `application.status` can sit at `publishing` for a while longer
  (the underlying container/route is starting async) before settling on
  `running` (live) or `error`. Poll `application.status`, not just
  `run.status`, to know when a deployment is actually live. `eds wf app
  delete` has also been observed returning a 500 while still mutating
  server-side state (`status` flips through `error`/`deleted` across
  repeated calls) -- treat it as unreliable/best-effort for now, not a
  guaranteed synchronous delete.
- `internal/config/config.go` — `Config` struct + `Load()`/`Save()`. Precedence
  is defaults -> `~/.config/eds/config.json` (or `$EDS_CONFIG` /
  `$XDG_CONFIG_HOME/eds/config.json`) -> `EDS_*` env vars -> CLI flags
  (flags applied later, in `cmd/helpers.go`). Config file is written with
  `0600` permissions and is a single platform-level file shared by both
  products (not split per-product), since project id is shared. Note: the
  internal `Config` struct field names and JSON keys (`api_url`, etc.) keep
  their pre-rename spelling — only the user-facing CLI flags/env vars/command
  names follow the `eds`/`repo`/`wf` product split; this is a deliberate
  minimal-diff choice, not an oversight.
- `internal/output/output.go` — `Printer` with `FormatAuto` (TTY -> table,
  pipe -> JSON), `FormatTable`, `FormatJSON`. All list/show commands render
  through `Printer.Table` / `Printer.KeyValue` / `Printer.PrintJSON` so both
  output modes stay in sync from one call site. Also has `HumanTime` /
  `HumanSize` formatters.
- `internal/repoapi/client.go` — generic authenticated HTTP client (`Do`),
  sends `X-API-KEY`, decodes JSON, maps non-2xx responses to
  `*APIError` (tries `error`, `message`, and per-field `errors` response
  shapes; also captures a server-assigned request id from common headers
  like `X-Request-Id` via `requestIDFromHeader` and includes it in
  `APIError.Error()` when present -- the single most useful thing to grab
  when reporting a bug to the API team, since bodies are sometimes empty
  even on a 500).
- `internal/repoapi/repositories.go` — repository-specific request/response
  types and the four API calls (`ListRepositories`, `GetRepository`,
  `CreateRepository`, `DeleteRepository`), all scoped under
  `/project/{projectID}/repository[ies]`. `ListRepositories` omits the
  `search` query param entirely when empty — the API treats an empty string
  as a literal (matches nothing) rather than a wildcard.
- `internal/workflow_client/` — generated OpenAPI client for the Workflow
  Studio API. `client.go` configures the HTTP client and sends `X-API-KEY`
  (the same key as Repo). `api_services.go`, `api_workflows.go`, and
  `api_pipelines.go` hold the application, run, and job API calls, all
  scoped under `/project/{projectID}/...`.
- `openapi-user.yaml` — OpenAPI (Swagger 2.0) spec for the Workflow Studio
  user API; the source of truth for request/response shapes when extending
  `internal/workflow_client`.
- `swagger/swagger.yaml` — OpenAPI spec for the upstream Repo API; the source
  of truth for request/response shapes when extending `internal/repoapi`.
  `swagger/swagger.ru.yaml` is a Russian-translated copy (all `summary`/
  `description` text values only — schema keys, types, `$ref`s untouched) for
  review purposes; keep it in sync by hand when `swagger.yaml` is
  re-synced from upstream.
- `scripts/install.sh` — self-installer (`curl | bash`), not part of the Go
  build. Downloads `eds-<os>-<arch>` assets from the latest GitHub Release of
  `cloud-ru/evolution-devservices-cli` by default and installs as `eds`;
  `EDS_CLI_BASE_URL`/`--base-url` points it at a custom mirror (e.g. the
  optional S3 upload from `make upload`) instead.
- `skill/SKILL.md` — Agent Skills description of the CLI's command surface,
  config precedence and error-handling contract for AI agents. Keep this in
  sync with `cmd/` when commands/flags change, since agents rely on it
  verbatim.

## Conventions specific to this codebase

- Every new subcommand should start with `ctx, err := resolveContext(cmd)`.
  If it hits a Repo-product API, call `ctx.requireAPIKey()`. If it hits a
  Workflow Studio API, call `ctx.ensureWorkflowAuth(cmd.Context())` — both
  validate that `EDS_API_KEY` is set.
- Every subcommand that returns structured data should support both table
  (default on TTY) and `--json` output via `ctx.Printer`, not ad-hoc
  `fmt.Println`.
- `eds repo` commands (and `eds wf app create --repository`) accept either a
  repository UUID or a name (`resolveRepoID` / `looksLikeUUID`); don't
  require callers to look up IDs first.
- New flags/env vars must follow the existing product-scoping convention:
  Repo-product settings get a `repo-`/`REPO_` infix (`--repo-api-url`,
  `EDS_REPO_API_URL`), Workflow Studio settings get a `wf-`/`WF_` infix
  (`--wf-api-url`, `EDS_WF_API_URL`), and only genuinely cross-product
  settings (project id and the API key) stay unprefixed. Don't add a third
  product's settings without picking an analogous short infix.
- Errors returned from `RunE` are printed as a single `Error: <msg>` line to
  stderr by `cmd.Execute()` (root sets `SilenceUsage`/`SilenceErrors`) — keep
  error messages one line and actionable (see `requireAPIKey`'s message as
  the template).
- Target platforms are Linux + macOS only (`DEFAULT_OSES := darwin linux` in
  the Makefile); don't add Windows-specific code paths without updating that.
- If you ever scaffold a `Dockerfile` for a Workflow Studio deploy test
  (fixtures, examples, docs), it must run as non-root with a read-only-ish
  root filesystem -- that's what the Container Apps runtime enforces.
  Plain `nginx:alpine` crash-loops there (`mkdir /var/cache/nginx/client_temp:
  Permission denied`); use `nginxinc/nginx-unprivileged:alpine` and
  `EXPOSE 8080` instead. Confirmed empirically end-to-end against prod.
- When the CLI's user-facing command surface changes (new command, new flag,
  changed defaults), update `README.md`'s Commands section and
  `skill/SKILL.md` in the same change — both are hand-maintained docs that
  will drift silently otherwise.
