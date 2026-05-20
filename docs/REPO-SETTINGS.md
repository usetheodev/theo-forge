# Repository Settings — Branch Protection & Supply Chain Posture

This document records the GitHub branch protection settings expected
on `main` (and `develop`). T5.9 / EC-7 from the
`fix-all-review-findings-plan`.

Branch protection is configured via GitHub UI / API, not committed
files; this document is the source of truth for what those settings
must be, so they can be audited and re-applied if the repo migrates.

## Branch: `main`

Apply via `gh api -X PUT /repos/usetheodev/theo-forge/branches/main/protection`
or the equivalent UI configuration:

- [ ] **Require a pull request before merging**
  - [ ] Require approvals: **1** minimum
  - [ ] Require review from Code Owners (uses `.github/CODEOWNERS`)
  - [ ] Dismiss stale pull request approvals when new commits are pushed
  - [ ] Require approval of the most recent reviewable push
- [ ] **Require status checks to pass before merging**
  - [ ] Require branches to be up to date before merging
  - [ ] Required checks: `Test`, `Lint`, `Security`
- [ ] **Require signed commits**
- [ ] **Require linear history**
- [ ] **Do not allow bypassing the above settings** (admins included)
- [ ] **Restrict who can push to matching branches** (maintainer team only)
- [ ] **Allow force pushes:** disabled
- [ ] **Allow deletions:** disabled

## Branch: `develop`

Same as `main`, plus:

- [ ] **Require linear history** — optional on develop, mandatory on main.

## Dependabot Settings

Dependabot PRs are subject to the same CODEOWNERS gate (see
`.github/CODEOWNERS`). This prevents a malicious Dependabot config or
manifest from auto-merging changes under `/.github/`.

If you enable auto-merge for Dependabot PRs elsewhere in the repo,
explicitly exclude `/.github/**` and `/go.mod` major-version bumps.

## Secret Scanning + Push Protection

- [ ] Secret scanning: **enabled**
- [ ] Push protection: **enabled**
- [ ] Dependency graph: **enabled**
- [ ] Dependabot security updates: **enabled**

## Verification

After applying the settings, attempt the following PRs in a sandbox fork:

1. Modify `.github/workflows/ci.yml` as a contributor — PR must require
   CODEOWNERS review.
2. Push a force update to main — must be rejected.
3. Submit a PR with a lint failure — `Lint` required check must block merge.

Save the JSON output of `gh api /repos/usetheodev/theo-forge/branches/main/protection`
to `docs/branch-protection-snapshot.json` for periodic audit.
