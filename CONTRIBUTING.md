# Contributing to eds

Thanks for considering a contribution to the Evolution DevServices CLI.

## Before you start

- For anything beyond a small fix, please open an issue first to discuss the
  change — it saves everyone time if the approach needs adjusting.
- By contributing, you agree that your contribution is submitted under the
  project's [Apache 2.0 license](LICENSE), per section 5 of that license.
- Please follow the [Code of Conduct](CODE_OF_CONDUCT.md) in all project
  spaces.

## Development setup

```bash
git clone git@github.com:cloud-ru/evolution-devservices-cli.git
cd evolution-devservices-cli
make build           # ./bin/eds
```

Go 1.22+ is required. Target platforms are Linux + macOS only.

Useful targets:

```bash
make lint
make test          # go test ./...
make build         # current platform into ./bin/
make build-all     # full matrix into ./dist/
```

Run `make lint` and `make test` before opening a PR — CI runs the same checks.

## Code conventions

See [CLAUDE.md](CLAUDE.md) for the architecture overview and codebase
conventions (command structure, config precedence, error message style,
product-scoped flag/env naming, etc.). It's kept up to date and is the best
starting point for understanding how a change should fit in.

In short:

- One command per file under `cmd/`, thin HTTP clients under `internal/`.
- Every subcommand that returns structured data outputs JSON by default
  (or YAML when `--yaml`) via `ctx.Printer`. `--quiet` suppresses all output.
- New flags/env vars follow the existing product-scoping convention
  (`--repo-*`/`EDS_REPO_*`, `--wf-*`/`EDS_WF_*`; only genuinely
  cross-product settings stay unprefixed).
- When the CLI's user-facing command surface changes, update `README.md`'s
  Commands section and `skills/eds/SKILL.md` in the same change.

## Submitting a change

1. Fork the repo and create a branch off `main`.
2. Make your change, with tests where it makes sense.
3. Run `make lint` and `make test`.
4. Open a pull request describing what changed and why.

## Reporting a security issue

Please do **not** open a public issue for a security vulnerability. Email
opensource@cloud.ru instead.

## Never commit secrets

Never commit API keys, tokens, or credentials. If you accidentally do, rotate
the credential immediately in addition to removing it from the branch.
