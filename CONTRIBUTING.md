# Contributing to TestPay

Thanks for your interest in contributing! This guide explains how we work and what we expect from contributors.

## Branching Strategy

We follow [GitHub Flow](https://docs.github.com/en/get-started/quickstart/github-flow) with a protected `main` branch.

- `main` is always deployable — every commit on `main` has passed CI and the 90% coverage gate.
- All work happens on short-lived branches cut from `main`.
- Branch names should use one of these prefixes:
  - `feature/<short-description>` — new features
  - `fix/<short-description>` — bug fixes
  - `chore/<short-description>` — tooling, docs, refactors
  - `hotfix/<short-description>` — urgent production fixes

Example: `feature/webhook-retry-backoff`, `fix/stripe-idempotency-key`.

## Pull Request Workflow

1. Fork the repo (external contributors) or create a branch (maintainers).
2. Make your changes. Keep commits focused and well-described.
3. Ensure all checks pass locally:
   ```bash
   make lint
   make test
   make coverage-check
   ```
4. Push your branch and open a PR against `main`.
5. Fill out the PR template completely.
6. Wait for CI to pass — all checks are required:
   - Lint (`go vet`)
   - Build
   - All tests (unit + integration) with `-race`
   - 90% coverage gate
   - Codecov upload
7. Address review feedback. Push additional commits; we squash-merge on merge.
8. A maintainer will squash-merge once approved and green.

## Code Style

- `gofmt` / `goimports` must be clean (CI enforces).
- Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Prefer small, focused interfaces.
- Error messages: lowercase, no trailing punctuation.
- Add tests for every new code path. The 90% coverage gate is enforced in CI.

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <short summary>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `perf`, `ci`, `build`.

## Releasing (Maintainers Only)

Releases are tag-driven. The release workflow runs automatically when a tag matching `v*.*.*` is pushed.

1. Ensure `main` is green and all desired changes are merged.
2. Update `CHANGELOG.md` if present.
3. Tag and push:
   ```bash
   git checkout main
   git pull
   git tag v0.1.0
   git push --tags
   ```
4. The release workflow will:
   - Build multi-arch Docker images
   - Publish to `ghcr.io/prashantluhar/testpay:v0.1.0` and `:latest`
   - Create a GitHub Release with auto-generated notes

## Maintainer Setup (one-time)

1. **Branch protection** — Go to **Settings → Branches → Add rule** for `main`:
   - Require a pull request before merging
   - Require approvals: 1
   - Require status checks to pass: `test`, `docker` (from CI workflow)
   - Require branches to be up to date before merging
   - Require conversation resolution before merging
   - Do not allow bypassing the above settings
2. **Secrets** — Settings → Secrets and variables → Actions:
   - `CODECOV_TOKEN` — from https://codecov.io after adding the repo
3. **Default branch** — verify `main` is the default branch
4. **Dependabot** — enabled automatically by `.github/dependabot.yml`
5. **CODEOWNERS** — enforced automatically when `.github/CODEOWNERS` exists

## Reporting Bugs

Open an issue with:
- A clear, reproducible test case
- Expected vs. actual behavior
- Version / commit SHA
- Relevant logs or output

## Security Issues

Please do NOT open a public issue for security vulnerabilities. Email the maintainer privately.

## Code of Conduct

Be kind. Assume good faith. We follow the [Contributor Covenant](https://www.contributor-covenant.org/).
