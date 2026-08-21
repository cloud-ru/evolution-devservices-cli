# eds — Evolution DevServices CLI for cloud.ru developer tools

`eds` is a small, agent-friendly command-line client for cloud.ru
developer tools products. It currently covers two products:

- **Repo** (`eds repo`) — git repositories in the cloud.
- **Workflow Studio** (`eds wf`) — wires a repository + branch to a deploy
  pipeline and publishes it.

Repo and Workflow Studio are independent products, but both are authenticated
with the same API key (`EDS_API_KEY`) via `X-API-KEY`. `eds` is just the
platform CLI they're both driven through. Every command is designed to be safely
driven by automation and AI agents: stable `--json` output, environment variables
for secrets.

- Single static binary (Go, no runtime dependencies).
- Reads the API key from `EDS_API_KEY` or from `~/.config/eds/config.json`.
- Outputs JSON when piped, pretty tables on a TTY (`--json` to force).
- Repo uses the local `git` CLI for clone. Both products authenticate with
  `X-API-KEY`. `eds repo clone` embeds the API key as HTTP Basic Auth
  credentials directly into the smart-HTTP URL it passes to `git clone`, so
  clone/push work standalone — no git credential helper, OS keychain, or
  `~/.netrc` needs to be pre-configured. This matters for CI and AI agent
  sandboxes, which typically have none of those.

## Installation

### From a GitHub Release (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/cloud-ru/evolution-devservices-cli/main/scripts/install.sh | bash
export PATH="$HOME/.local/bin:$PATH"
eds version
```

The installer detects the platform, downloads the matching static binary
from the latest [GitHub Release](https://github.com/cloud-ru/evolution-devservices-cli/releases),
places it in `~/.local/bin/eds`, and verifies the install.

You can also grab a binary directly from the
[releases page](https://github.com/cloud-ru/evolution-devservices-cli/releases/latest):

```bash
# pick eds-<os>-<arch> for your platform, e.g. eds-darwin-arm64, eds-linux-amd64
curl -fsSL -o eds \
  https://github.com/cloud-ru/evolution-devservices-cli/releases/latest/download/eds-darwin-arm64
chmod +x eds && sudo mv eds /usr/local/bin/eds
```

### From source

```bash
git clone git@github.com:cloud-ru/evolution-devservices-cli.git
cd evolution-devservices-cli
go install github.com/cloud-ru/evolution-devservices-cli@latest
# or
make build           # ./bin/eds
make build-all       # cross-compile darwin/linux × amd64/arm64 into ./dist/
```

> Target platforms: **Linux + macOS** (developers locally + CI).
> Override the matrix if you ever need to: `make build-all OSES="linux darwin" ARCHS="amd64 arm64"`.

## Configuration

The CLI looks for configuration in this order (later wins):

1. Built-in defaults: Repo API URL `https://devtools.api.cloud.ru/repo/api/v1`,
   git host `https://repo.cloud.ru/`, Workflow Studio API URL
   `https://pipeline.cloud.ru/public-api/v1`.
2. File `~/.config/eds/config.json` (overridable via `EDS_CONFIG` or
   `XDG_CONFIG_HOME`) — a single platform-level file shared by both products.
3. Environment variables (see table below).
4. Per-command flags.

Env vars and flags are namespaced by product for URLs — `EDS_REPO_*`/`--repo-*`
for Repo, `EDS_WF_*`/`--wf-*` for Workflow Studio — except for `--api-key`/
`EDS_API_KEY` and `--project`/`EDS_PROJECT_ID`, which are shared platform-level
settings.

Example config file (`~/.config/eds/config.json`):

```json
{
  "api_url": "https://devtools.api.cloud.ru/repo/api/v1",
  "project_id": "3232b2d0-1063-41e6-b2fa-13df767f4a0a",
  "api_key": "...",
  "git_host": "https://repo.cloud.ru/",
  "workflow_api_url": "https://pipeline.cloud.ru/public-api/v1"
}
```

### Production vs dev (Repo product)

- **Production** — `https://devtools.api.cloud.ru/repo/api/v1`, git host
  `https://repo.cloud.ru/`.
- **Dev** — `https://devtools.dev.api.internal.cloud.ru/repo/api/v1`.

Override with `--repo-api-url` or `EDS_REPO_API_URL`.

### Workflow Studio auth

Workflow Studio uses the same API key as Repo (`--api-key` / `EDS_API_KEY`),
sent as `X-API-KEY` on every request. There is no separate credential pair or
token exchange — the same key works for both `eds repo *` and `eds wf *`.

## Commands

```text
eds login                                  save credentials to the local config
eds config                                 show effective configuration
eds version                                print the CLI version

eds repo list [--search S] [--sort ...]    list git repositories
eds repo create <name>                     create a repository
eds repo show <id-or-name>                 show repository details
eds repo delete <id-or-name> [--force]     delete a repository
eds repo clone <id-or-name> [dir] [--ssh]  clone via local git CLI
eds repo remote-add <id-or-name> [--name N] [--ssh]  wire an existing local checkout to it (git remote add)

eds wf app create <name> --repository R|--repository-url URL --branch B   create and auto-deploy a Workflow Studio application
eds wf app list                                       list applications
eds wf app show <id>                                  show application details
eds wf app update <id> --branch B [--name N]          update name/branch
eds wf app delete <id> [--force]                       delete an application
eds wf app deploy <id>                                run the pipeline and publish
eds wf app deployments <id>                           list publish history

eds wf run show <id>                          show a run's status, stages and jobs
eds wf run stop <id>                          stop a running run

eds wf job logs <id>                          get logs for a job
```

### Login

```bash
eds login --api-key "$EDS_API_KEY" --project <project-id>
eds login --repo-api-url https://devtools.dev.api.internal.cloud.ru/repo/api/v1 \
          --api-key "$EDS_API_KEY" --project <project-id>
```

### For an AI agent

The CLI is designed to be safely scripted. Recommended pattern:

```bash
export EDS_API_KEY="..."
export EDS_PROJECT_ID="..."

# List repos as JSON, pipe into jq
eds repo list --json | jq '.repositories[].name'

# Create a repo, capture the new id
eds repo create demo --json | jq -r '.id'

# Inspect
eds repo show demo --json | jq '.clone.https'

# Clone (uses the system's git)
eds repo clone demo ./work/demo
```

All commands return exit code 0 on success, non-zero on failure, and
write errors to stderr in a single line:

```
Error: api error 404: {"error_msg":"404 Not Found"}
```

See [`skill/SKILL.md`](skill/SKILL.md) for the canonical Agent Skills
description that can be attached to an AI agent.

## Workflow Studio (deploy & publish)

Workflow Studio is a "developer tools" product on cloud.ru: it wires a
repository + branch to a deploy pipeline. An **application** is that wiring;
running the application's pipeline (a **deployment**) publishes it and
produces a live URL. This is the CLI's main "vibe-coded a site, now ship it"
path — `eds repo` gets your code hosted, `eds wf app` publishes it.

```bash
export EDS_API_KEY="..."
export EDS_PROJECT_ID="..."

# 1. Create the repo and push code (see "File upload / push" below)
eds repo create my-site
eds repo clone my-site && cd my-site
# ...add code + Dockerfile...
git add . && git commit -m "init" && git push origin main

# 2. Wire it to a Workflow Studio application. Creating it auto-triggers the
#    first deploy — you don't need a separate `deploy` call right after create.
APP_ID=$(eds wf app create my-site --repository "$REPO_ID" --branch main --json | jq -r '.id')

# 3. Poll until it's actually live. Check application.status (not just run
#    status) because "publishing" can last well after the run already reports
#    "done".
for i in $(seq 1 20); do
  STATUS=$(eds wf app show "$APP_ID" --json | jq -r '.status')
  echo "status: $STATUS"
  [ "$STATUS" = "running" ] || [ "$STATUS" = "error" ] && break
  sleep 15
done
URL=$(eds wf app deployments "$APP_ID" --json | jq -r '.deployments[0].url')
echo "url: $URL"
```

If a deployment fails, drill into the failing job's logs:

```bash
eds wf app show "$APP_ID" --json | jq -r '.run.stages[].jobs[] | select(.status=="failed") | .id' \
  | xargs -I{} eds wf job logs {}
```

`eds wf run` and `eds wf job` are the lower-level primitives behind `eds wf
app show` — use them directly when you need to inspect or control a
particular run (e.g. `eds wf run stop`) or stream job logs (`eds wf job logs`).

## File upload / push

This CLI does **not** implement a custom upload path. To push code, use
the standard git workflow after `eds repo clone`:

```bash
eds repo clone demo
cd demo
git add . && git commit -m "init" && git push origin main
```

`git push` works out of the box because `eds repo clone` already embedded the
API key as basic-auth credentials in `origin`'s URL (see `.git/config`) — no
git credential helper or OS keychain is involved. Note this means the API
key sits in plaintext in that repo's `.git/config`; treat the clone
directory with the same care as the key itself.

If the code already exists locally (no `eds repo clone` involved) and you
just created the remote repository, use `eds repo remote-add` instead of a
plain `git remote add` to get the same embedded authentication:

```bash
eds repo create demo
cd path/to/existing/local/repo
eds repo remote-add demo
git push -u origin main
```

`--ssh` is the one exception: it depends on the host's SSH public key being
registered separately (not handled by this CLI), so it still needs whatever
ambient SSH setup the environment provides.

## Distribution / publishing

Releases are published on
[GitHub Releases](https://github.com/cloud-ru/evolution-devservices-cli/releases).

```bash
# cross-compile everything into ./dist/ + sha256 checksums
make release

# tag, then cut the release with the built artifacts
git tag v0.2.0 && git push origin v0.2.0
gh release create v0.2.0 dist/* --generate-notes
```

`scripts/install.sh` downloads from the latest GitHub Release by default.
The `Makefile` also has `make upload`/`make upload-latest` targets for
optionally mirroring builds to an S3-compatible bucket (e.g. for internal
environments without GitHub access) — pass `EDS_CLI_BASE_URL` to
`install.sh` to install from a mirror instead of GitHub:

```bash
export AWS_ENDPOINT_URL=https://storage.cloud.ru
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
make upload BUCKET=my-bucket PREFIX=evolution-devservices-cli VERSION=v0.2.0

curl -fsSL https://storage.cloud.ru/my-bucket/evolution-devservices-cli/install.sh | \
  EDS_CLI_BASE_URL=https://storage.cloud.ru/my-bucket/evolution-devservices-cli bash
```

## Development

```bash
make lint
make test          # go test ./...
make build         # current platform into ./bin/
make build-all     # full matrix into ./dist/
make clean         # remove ./bin and ./dist
make help          # list targets
```

Project layout:

```
main.go                         # entry point + ldflags-driven version
cmd/
  root.go                       # cobra root command (`eds`) + global flags
  helpers.go                    # config resolution, runtime context
  login.go                      # `eds login`
  config.go                     # `eds config`
  version.go                    # `eds version`
  repo.go                       # `eds repo list|create|show|delete|clone`
  wf.go                         # `eds wf` parent command (groups app/run/job)
  app.go                        # `eds wf app create|list|show|update|delete|deploy|deployments`
  run.go                        # `eds wf run show|stop`
  job.go                        # `eds wf job logs`
internal/
  config/                       # disk config + env overrides
  output/                       # JSON / table formatting
  repoapi/                      # thin HTTP client for the Repo product API
  workflow_client/              # generated OpenAPI client for Workflow Studio
scripts/
  install.sh                    # one-liner installer (GitHub Releases by default)
skill/
  SKILL.md                      # Agent Skills description
Makefile                        # build, build-all, release, upload, …
```

## Environment variables

| Variable          | Description                                                 |
| ----------------- | ------------------------------------------------------------ |
| `EDS_PROJECT_ID`  | Default project ID (shared across products)                  |
| `EDS_REPO_API_URL`  | Repo product API base URL (overrides config)                |
| `EDS_API_KEY`       | API key used for both products (X-API-KEY)                   |
| `EDS_REPO_GIT_HOST` | Repo product git smart-HTTP host                             |
| `EDS_WF_API_URL`  | Workflow Studio product API base URL (overrides config)      |
| `EDS_CONFIG`      | Path to config file (overrides default)                      |
| `XDG_CONFIG_HOME` | Respected when locating the config file                      |
| `AWS_ENDPOINT_URL` | S3-compatible endpoint for `make upload` (optional mirror)  |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development setup and PR
process, and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community
guidelines. Report security issues to opensource@cloud.ru rather than a
public issue.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
