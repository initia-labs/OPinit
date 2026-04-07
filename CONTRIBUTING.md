# Contributing

This repository uses Conventional Commits for commit messages and pull request titles.
Before opening a PR, make sure your changes match the repository's existing Go,
protobuf, and upgrade workflows.

## Commit Convention

Use this format for commits:

```text
type(scope): subject
```

Examples:

```text
feat(opchild): add MsgRelayOracleData to system lane
fix(ophost): handle empty batch submission
docs(readme): update build instructions
test(opchild): cover validator query pagination
chore(deps): bump SDK dependency
```

Rules:

- use lowercase `type` and `scope`
- keep the subject short and imperative
- do not end the subject with a period
- if the change is breaking, add `!` in the Conventional Commit prefix
  (for example: `feat(opchild)!: change oracle data response`)

Common types used in this repository:

- `feat`
- `fix`
- `docs`
- `refactor`
- `test`
- `chore`
- `build`
- `ci`

## Branch Naming

Use a short branch name that describes the change.

Format:

```text
type/short-description
```

Examples:

```text
fix/ophost-batch-submission
feat/opchild-system-lane
docs/contributing-guide
test/opchild-query-pagination
```

Keep branch names focused. If a branch mixes unrelated work, split it before
opening a PR.

## Pull Requests

PR titles should follow the same Conventional Commit format:

```text
type(scope): subject
```

For breaking changes, add `!` in the PR title prefix:

```text
type(scope)!: subject
```

Examples:

```text
fix(ophost): handle empty batch submission
feat(opchild): add MsgRelayOracleData to system lane
test(opchild): cover query edge cases
feat(ophost)!: rename bridge config fields
```

Follow the PR template in `.github/PULL_REQUEST_TEMPLATE.md`.
At minimum, every PR should clearly describe:

- what changed
- why it changed
- how it was validated
- whether the change is breaking

If the change is tied to an issue, proposal, or spec, link it in the PR body.

## Validation

Run the smallest relevant validation set for your change before pushing.

Common commands:

- focused package tests: `go test ./path/to/package -run <TestName> -count=1`
- full unit suite: `make test-unit`
- race-enabled suite: `make test-race`
- coverage run: `make test-cover`
- lint: `make lint`

Prefer focused tests while iterating, then run the broader repository command
that matches the risk of the change.

If a change affects cross-module behavior, add or run at least one regression
test that exercises the full path.

## Formatting and Generated Files

When you change Go code, keep formatting and imports clean:

- `make format` (if available) or ensure `golangci-lint` passes

When you change protobuf definitions, regenerate the related outputs:

- `make proto-gen`
- `make proto-pulsar-gen`

For protobuf changes, also make sure lint and breaking-change checks still make
sense for the PR:

- `make proto-lint`

If generated protobuf outputs change, include those generated files in
the same PR.

## Scope Discipline

Keep each PR focused on one logical change. Avoid mixing unrelated refactors,
formatting-only edits, generated-file churn, and behavior changes unless they
must land together.
