---
name: evolution-devservices-cli
description: Manage cloud.ru developer tools products via the `eds` CLI — git repositories (Repo product, "eds repo") and deploy/publish pipelines (Workflow Studio product, "eds wf"). Use when the user asks to list, create, inspect, delete or clone repositories, to push/pull code, or to create/deploy/monitor a Workflow Studio application.
---

# Evolution DevServices CLI (eds) — Agent Skill

The `eds` CLI is a thin, agent-friendly wrapper around two independent
cloud.ru "developer tools" products:

- **Repo** (`eds repo`) — git repositories.
- **Workflow Studio** (`eds wf`) — deploy pipelines / publishing.

Both products are authenticated with the same API key (`EDS_API_KEY`).

It produces stable JSON output, accepts configuration through environment
variables, and is safe to invoke from automation.

This skill teaches an AI agent how to drive the CLI for the common
tasks: repository discovery/creation/inspection/deletion/cloning, and
Workflow Studio application creation/deployment/monitoring.

## When to use this skill

Reach for `eds repo *` whenever the user wants to:

- find a repository by name or list all repositories in a project,
- create a new git repository,
- check whether a repository exists or read its metadata,
- delete a repository,
- clone a repository locally (for example, to inspect files or push code).

Reach for `eds wf app *` / `eds wf run *` / `eds wf job *` whenever the
user wants to:

- **publish or deploy** a repository as a live service (the main scenario:
  code was written and pushed to a Repo repository, now it needs to go live),
- check the status of a deployment/publish (run/stage/job progress, the
  live URL once it succeeds),
- inspect or control a specific pipeline run or job (stop, retry, read logs).

Use `eds config` to inspect the active configuration / project id.

Do **not** use `eds` for non-git, non-deploy operations (model cards,
datasets, merge requests). Those are out of scope for this CLI.

## Tool contract

Every command supports `--json` for machine-readable output. When the
output is piped to another command, JSON is selected automatically.
Errors go to stderr and the process exits non-zero.

| Command                                                                                                                                 | Purpose                                                                                                                                                     |
|-----------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `eds version`                                                                                                                           | Print the installed CLI version                                                                                                                             |
| `eds config`                                                                                                                            | Print effective configuration                                                                                                                               |
| `eds repo list [--search S] [--sort name_asc\|name_desc\|updated_at_asc\|updated_at_desc] [--limit N] [--offset N] [--json]`            | List repositories in the configured project                                                                                                                 |
| `eds repo create <name> [--description "..."] [--visibility private\|shadow] [--json]`                                                  | Create a new git repository                                                                                                                                 |
| `eds repo show <id-or-name> [--json]`                                                                                                   | Show details: id, default_branch, size, clone URLs                                                                                                          |
| `eds repo delete <id-or-name> [--force] [--json]`                                                                                       | Delete (irreversible; requires confirmation unless `--force`)                                                                                               |
| `eds repo clone <id-or-name> [dir] [--ssh] [--target DIR]`                                                                              | Clone via local `git` CLI                                                                                                                                   |
| `eds wf app create <name> --repository R\|--repository-url URL --branch B [--json]`                                                     | Create a Workflow Studio application from a repo + branch                                                                                                   |
| `eds wf app list [--search S] [--sort created_at_asc\|created_at_desc] [--json]`                                                        | List applications                                                                                                                                           |
| `eds wf app show <id> [--json]`                                                                                                         | Show application details (status, run_id, pipeline_id, ...)                                                                                                 |
| `eds wf app update <id> --branch B [--name N] [--json]`                                                                                 | Update an application's name/branch                                                                                                                         |
| `eds wf app delete <id> [--force] [--json]`                                                                                             | Delete an application and its deployments (irreversible)                                                                                                    |
| `eds wf app deploy <id> [--json]`                                                                                                       | Run the pipeline and publish (the "deploy" action)                                                                                                          |
| `eds wf app deployments <id> [--json]`                                                                                                  | List publish history for an application                                                                                                                     |
| `eds wf run show <id> [--json]` / `eds wf run stop <id>`                                | Inspect/control a pipeline run                                                                                                                              |
| `eds wf job logs <id>`                                                                  | Stream job logs                                                                                                                                             |
| `eds login --api-key <KEY> --project <ID> [--repo-api-url URL]`                                                                         | Persist credentials (one-time setup)                                                                                                                      |

The `<id-or-name>` argument on `eds repo *` and `--repository` on
`eds wf app create` are resolved automatically: UUIDs are used as-is, names
are looked up against the configured project. All other `<id>` arguments
(application, run, job) are Workflow Studio ids returned by a previous call
and must be passed as-is.

## Configuration

`eds repo *` and `eds wf *` share the same API key (`EDS_API_KEY`).
Flags/env vars are namespaced by product for URLs: `--repo-*`/`EDS_REPO_*`
for Repo, `--wf-*`/`EDS_WF_*` for Workflow Studio; `--api-key`/`EDS_API_KEY`
and `--project`/`EDS_PROJECT_ID` are shared platform-level settings.

Precedence (lowest → highest):

1. Built-in defaults: `api_url=https://devtools.api.cloud.ru/repo/api/v1`,
   `git_host=https://repo.cloud.ru/`,
   `workflow_api_url=https://pipeline.cloud.ru/public-api/v1`.
2. File at `~/.config/eds/config.json` (override with `EDS_CONFIG`).
3. Environment: `EDS_PROJECT_ID`, `EDS_REPO_API_URL`,
   `EDS_API_KEY`, `EDS_REPO_GIT_HOST`, `EDS_WF_API_URL`.
4. Flags: `--project`, `--repo-api-url`, `--api-key`,
   `--repo-git-host`, `--wf-api-url`.

Dev environment: `EDS_REPO_API_URL=https://devtools.dev.api.internal.cloud.ru/repo/api/v1`.

## Required environment for an agent

Before invoking any `eds repo *` or `eds wf *` command, the agent must
ensure:

- `EDS_API_KEY` is set (or the key was saved via `eds login --api-key`).
- `EDS_PROJECT_ID` is set.

The skill's runtime should arrange for these before the first call.

## Installation inside the agent's sandbox

```bash
curl -fsSL https://raw.githubusercontent.com/cloud-ru/evolution-devservices-cli/main/scripts/install.sh | bash
export PATH="$HOME/.local/bin:$PATH"
eds version   # smoke-test
```

The installer detects the platform (darwin/linux × amd64/arm64),
downloads the matching binary into `~/.local/bin/eds`, and verifies it.

Supported platforms: **Linux + macOS** (developers locally + CI).

## Recipes

### Discover: list all repositories

```bash
eds repo list --json | jq '.repositories[] | {id, name, visibility: .visibility_level}'
```

### Search by substring

```bash
eds repo list --search demo --json
```

### Find a repository by exact name

```bash
eds repo list --search my-repo --json | \
  jq -r '.repositories[] | select(.name == "my-repo") | .id'
```

### Create a new repository and remember its id

```bash
NEW_ID=$(eds repo create my-new-repo --description "agent created" --json | jq -r '.id')
echo "Created repository id=$NEW_ID"
```

### Check existence + default branch before cloning

```bash
eds repo show my-new-repo --json | jq '{id, default_branch, clone: .clone.https}'
```

### Clone to a specific directory

```bash
eds repo clone my-new-repo ./work/my-new-repo
cd ./work/my-new-repo
git status
```

### Push code (after clone)

The CLI does not implement a custom upload path — use git directly:

```bash
cd ./work/my-new-repo
git add . && git commit -m "init" && git push origin main
```

### Delete a repository (with confirmation)

```bash
echo "my-old-repo" | eds repo delete my-old-repo --force --json
```

### Inspect the active configuration

```bash
eds config --json
# { "project_id": "...",
#   "repo_api_url": "...", "api_key": "abcd…wxyz", "repo_git_host": "...",
#   "wf_api_url": "..." }
```

### Publish a repository as a live service (main Workflow Studio scenario)

This is the end-to-end path from "code was pushed to a repository" to
"it's live at a URL" — the primary reason to use `eds wf app *`:

```bash
export EDS_API_KEY=...
export EDS_PROJECT_ID=...

# 1. The repository already exists (created + pushed via `eds repo`), and its
#    Dockerfile is Container Apps-compatible -- see "Dockerfile requirements"
#    below. Skipping that step is the #1 cause of a deploy that never goes live.
REPO_ID=$(eds repo list --search my-site --json | jq -r '.repositories[] | select(.name=="my-site") | .id')

# 2. Wire it to a Workflow Studio application. Creating it auto-triggers the
#    first deploy -- you don't need a separate `deploy` call right after create.
APP_ID=$(eds wf app create my-site --repository "$REPO_ID" --branch main --json | jq -r '.id')

# 3. Poll until it's actually live. IMPORTANT: check application.status, not
#    just the run status -- application.status goes for_create -> publishing
#    -> running (live) or error, and "publishing" can last well after the
#    underlying run already reports "done" (the container is still starting).
for i in $(seq 1 20); do
  STATUS=$(eds wf app show "$APP_ID" --json | jq -r '.status')
  echo "status: $STATUS"
  [ "$STATUS" = "running" ] || [ "$STATUS" = "error" ] && break
  sleep 15
done
eds wf app show "$APP_ID" --json | jq '{status: .status, url: .repository_url}'
```

To redeploy later (e.g. after a new push), skip straight to `eds wf app
deploy "$APP_ID"` — the application already exists.

### Dockerfile requirements for Workflow Studio (Container Apps)

The runtime executes containers as **non-root with a read-only-ish root
filesystem**. A plain `nginx:alpine` Dockerfile crash-loops there:

```
nginx: [emerg] mkdir() "/var/cache/nginx/client_temp" failed (13: Permission denied)
```

Use the unprivileged image and a non-privileged port instead:

```dockerfile
FROM nginxinc/nginx-unprivileged:alpine
COPY index.html /usr/share/nginx/html/index.html
EXPOSE 8080
```

Confirmed working end-to-end against prod. If you're scaffolding a
Dockerfile for any other base image, assume the same constraint (no root,
no writing outside a few known-writable paths) and pick a variant/config
built for unprivileged operation.

### Debug a failed deployment

```bash
eds wf app show "$APP_ID" --json | \
  jq -r '.run.stages[].jobs[] | select(.status=="failed") | .id' | \
  while read -r JOB_ID; do eds wf job logs "$JOB_ID"; done
```

### Redeploy an existing application (e.g. after a new push)

```bash
eds wf app deploy "$APP_ID" --json | jq -r '.run_id'
```

## Error handling

The CLI prints a single-line error to stderr and exits with a non-zero
status on failure. Suggested agent policy:

- **2xx**: parse stdout as JSON (when `--json`) or treat stdout as human-readable text.
- **non-zero**: read stderr, retry only on transient errors (timeouts, 5xx). Do **not**
  retry on 4xx — they indicate an agent bug (wrong id, missing key, etc.).

Sample failure:

```
$ eds repo list
Error: API key is not set. Run `eds login --api-key <KEY>` or set EDS_API_KEY
exit=1
```

## Limits and non-features

- Only `type=git` repositories are created (model/dataset are out of scope).
- `--ssh` requires the host's SSH public key to be registered separately
  (not handled by this CLI).
- No custom upload mechanism: use `git push` after `eds repo clone`.
- Workflow Studio application creation assumes a **default environment**
  already exists in the project; the CLI does not manage environments.
- `eds wf job logs` streams the server's SSE log feed as plain lines — it
  does not (yet) expose structured log levels or timestamps beyond what
  the server sends.
- The default Dockerfile pattern most agents reach for (`nginx:alpine`)
  does **not** work on Workflow Studio's runtime — see "Dockerfile
  requirements" above. Always use `nginxinc/nginx-unprivileged` (or another
  non-root-friendly base) for anything deployed via `eds wf app`.
- `eds wf app delete` has been observed returning `workflow api error 500`
  while still changing the application's state server-side (bouncing
  between `error`/`deleted` across repeated calls). Treat it as best-effort:
  call it, then check `eds wf app list`/`show` to see the actual outcome
  rather than trusting the delete call's own exit code.

## Quick reference

```text
# 1. bootstrap
export EDS_API_KEY=...
export EDS_PROJECT_ID=...
# (optional) install via curl | bash — see "Installation"

# 2. read
eds repo list --json | jq '.repositories[].name'

# 3. write
NEW_REPO=$(eds repo create my-app --json | jq -r '.id')

# 4. deploy
APP_ID=$(eds wf app create my-app --repository "$NEW_REPO" --branch main --json | jq -r '.id')
eds wf app show "$APP_ID" --json | jq '{status: .status, url: .repository_url}'
```
