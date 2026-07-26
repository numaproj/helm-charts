# Helm Chart Release Playbook

Step-by-step instructions for cutting a new `numaproj/numaflow` Helm chart release in this repo.

Every step is independently runnable and produces output a maintainer can review before moving on. Inline commands assume the new numaflow version is **`v1.8.0`** — substitute as needed for future releases.

> **Quick reference:** see [MAINTAINERS.md](./MAINTAINERS.md) for the condensed summary. This document is the full walkthrough.

---

## What gets automated, what doesn't

| Step                                                              | Automated by                                       |
| ----------------------------------------------------------------- | -------------------------------------------------- |
| `Chart.yaml` `version` + `appVersion` bump                        | `make upgrade-charts`                              |
| `values.yaml` image `tag`                                         | `make upgrade-charts`                              |
| CRDs, RBAC, ServiceAccounts                                       | `make upgrade-charts`                              |
| Configmaps, Deployments, Secrets, Services (templated mirror)     | `make sync-mirrored-files`                         |
| Merge-conflict resolution between upstream changes and chart code | **Manual** — review `upgrade/.merge-rejects/`      |
| `README.md` version-compatibility table                           | **Manual** — add one row                           |
| Local install verification                                        | **Manual** — `helm lint` / `helm template` / `helm install` |
| Opening + reviewing the PR                                        | **Manual**                                         |
| Publishing the chart to GitHub Releases / `gh-pages`              | CI on merge to `main` (chart-releaser-action)      |

---

## Prerequisites

- Working tree on `main`, clean and up to date with `origin/main`.
- Numaflow release tag (e.g. `v1.8.0`) is published at <https://github.com/numaproj/numaflow/releases> — the tool will fail fast if it isn't.
- Local tools on `PATH`: `git`, `diff`, `patch`, `go` (>= 1.24), `helm` (>= 3.14).
- Optional: a local k8s cluster (`kind` or `minikube`) for the install check in Step 6. If you don't have one, CI's `lint-test` job will cover install verification on the PR.

```bash
# Sanity check
which git diff patch go helm
git status                       # expect: clean
git pull --ff-only origin main
```

---

## Step 0 — create a release branch

```bash
git checkout -b prep-v1.8.0-release
```

This matches the pattern from prior releases (commits `ce48a7b`, `da10eed`, `bb33931`). The branch name encodes the numaflow version, not the chart version, so it stays readable across major/minor/patch bumps.

---

## Step 1 — refresh the auto-managed files

```bash
NUMAFLOW_VERSION=v1.8.0 make upgrade-charts
```

What this does:

1. Verifies `v1.8.0` exists on `github.com/numaproj/numaflow` via the GitHub releases API. If the tag is missing, you get a clear `Version check failed` error and the tool exits — nothing has been modified yet.
2. Rewrites `charts/numaflow/Chart.yaml`:
   - `appVersion` → `1.8.0`
   - `version` → bumped per the chart's own semver rule (major numaflow bump → chart major, minor numaflow bump → chart minor, patch numaflow bump → chart patch). For `1.7.5 → 1.8.0` (minor bump) the chart version goes from `0.4.6 → 0.5.0`.
3. Rewrites `charts/numaflow/values.yaml`: image `tag: v1.8.0`.
4. Pulls fresh CRDs into `charts/numaflow/crds/` (5 files).
5. Pulls fresh RBAC into `charts/numaflow/templates/rbac/{cluster-scoped,namespaced}/` (17 files).
6. Pulls fresh ServiceAccounts into `charts/numaflow/templates/serviceaccounts/` (4 files).

**After running, review:**

```bash
git status
git diff --stat
git diff charts/numaflow/Chart.yaml charts/numaflow/values.yaml
```

The Chart.yaml/values.yaml diff should be tiny (3 lines). RBAC/CRD/SA files may show whitespace-only or trailing-newline diffs even in releases where upstream didn't change them — visually scan each before moving on.

---

## Step 2 — dry-run the mirror sync

```bash
NUMAFLOW_VERSION=v1.8.0 make sync-mirrored-files
```

What this does (per file, for all 15 mirrored files):

1. Fetches the upstream blob at `v1.7.5` (chart's previous `appVersion`) **and** at `v1.8.0`.
2. If both upstream blobs are byte-identical → `no-change`, nothing to do.
3. Otherwise, runs a 3-way merge:
   - **base** = upstream@v1.7.5
   - **theirs** = upstream@v1.8.0
   - **ours** = chart's templated copy (Helm expressions are tokenized for the merge then restored)
4. Writes the result to `upgrade/.merge-rejects/<filename>.merged` (clean) or `<filename>.conflict` (with `<<<<<<< / ||||||| / ======= / >>>>>>>` markers).

By default the run is **dry** — clean merges are not applied to the chart files. The summary table at the end shows the per-file outcome and the exit code is non-zero if any file ended in conflict.

```text
Mirror sync summary
===================
  no-change:        N
  applied:          0
  ready:            M     ← clean merges waiting to be applied
  conflict:         K     ← review and resolve manually
  upstream-missing: 0
  errors:           0
```

**Flags you may want (pass via `SYNC_FLAGS=...`):**

- `--from-version=vX.Y.Z` — override the baseline. Useful when catching up multiple releases at once (e.g. you skipped `v1.7.6`).
- `--only=numaflow-controller.yaml,numaflow-server.yaml` — restrict the run to specific basenames for incremental work.

---

## Step 3 — review the dry run

```bash
ls upgrade/.merge-rejects/
```

For each entry in that directory:

### `*.merged` — clean merges proposed by the tool

```bash
diff -u charts/numaflow/templates/<path-from-registry>/<file>.yaml \
        upgrade/.merge-rejects/<file>.merged
```

Verify the diff matches your expectation of what upstream changed. The tool preserves Helm templating (`{{ ... }}`, `{{- if }}` blocks, the labels include, the `{{ .Release.Namespace }}` line) — anything outside those should reflect upstream's edits.

### `*.conflict` — files needing manual resolution

```bash
less upgrade/.merge-rejects/<file>.conflict
```

Conflict regions look like:

```text
<<<<<<< ours
   <what the chart currently has (with templates restored)>
||||||| base
   <what upstream had at the previous version>
=======
   <what upstream has at the new version>
>>>>>>> theirs
```

Decide per region:

- **Take theirs** if the chart was just mirroring upstream verbatim and the upstream change is correct for the chart too.
- **Take ours** if the chart deliberately diverges (e.g. a key was templated to support a `--set` override that upstream's literal default doesn't capture).
- **Hybrid** if the answer is "both" — for example, accept upstream's added key but keep our existing templated key untouched.

To resolve:
1. Edit the `.conflict` file, deleting the marker lines and choosing/blending content as above.
2. Copy the resolved content over the corresponding chart file:
   ```bash
   cp upgrade/.merge-rejects/<file>.conflict charts/numaflow/templates/<dir>/<file>.yaml
   ```
3. Delete the rejects file:
   ```bash
   rm upgrade/.merge-rejects/<file>.conflict
   ```

**Known limitation — adjacency false positives:** when upstream changes a line that sits within ~3 rows of a line the chart has templated, git's xdiff bundles both into one hunk and reports a conflict even though the changes don't actually overlap. The "Take ours" half of the conflict region is the chart's existing line; the "Take theirs" half is upstream's change to a *different* line. Both can be accepted; the conflict just couldn't be auto-resolved because the hunks were too close together. See `TestThreeWayMerge_AdjacencyLimitation` in `upgrade/internal/mirror/threeway_test.go` for the canonical example.

---

## Step 4 — apply the clean merges

Once the dry-run output looks right and any conflicts have been resolved by hand:

```bash
NUMAFLOW_VERSION=v1.8.0 make sync-mirrored-files SYNC_FLAGS=--apply
```

This re-runs the sync, but this time clean merges are written back to the chart files. Files you already resolved by hand in Step 3 should now report `no-change` (because your hand-resolved content matches what upstream wants, or because you accepted upstream's version directly).

If any file still reports `conflict` after Step 3, you missed one — go back, resolve, repeat.

```bash
git diff --stat charts/numaflow/templates/
```

Should show small focused changes; nothing in `crds/`, `rbac/`, or `serviceaccounts/` here (those were handled in Step 1).

---

## Step 5 — update `README.md`

The version-compatibility table at the repo root is not auto-updated. Add a new row at the top of the table (just below the column headers, above the existing `0.4.5 / 1.7.5` row):

```text
numaproj/numaflow       0.5.0           1.8.0           A Helm chart for installing Numaflow in Kubernetes
```

Use the chart version that the tool wrote to `Chart.yaml` in Step 1 (`grep '^version:' charts/numaflow/Chart.yaml`) — it may not always be the value you guessed.

```bash
git diff README.md
```

---

## Step 6 — local verification

### 6a. Lint

```bash
helm lint charts/numaflow
```

Should print `0 chart(s) failed`.

### 6b. Render

```bash
helm template charts/numaflow --namespace numaflow-system | head -80
helm template charts/numaflow --namespace numaflow-system \
  | grep -E 'image:|appVersion'
```

Verify:
- No template errors.
- Controller, server, dex-server, and webhook deployments all reference `quay.io/numaproj/numaflow:v1.8.0` (or `numaflow-server:v1.8.0` where appropriate).

### 6c. Install in a local cluster (optional but recommended)

```bash
helm install numaflow charts/numaflow \
  --namespace numaflow-system --create-namespace
kubectl -n numaflow-system get pods
kubectl -n numaflow-system get deploy numaflow-controller \
  -o jsonpath='{.spec.template.spec.containers[0].image}'
helm uninstall numaflow -n numaflow-system
kubectl delete ns numaflow-system
```

All pods should reach `Running`. If you can't run a cluster locally, the PR's `lint-test` CI job exercises the same path on a KinD cluster.

---

## Step 7 — open the PR

```bash
git status                       # confirm: only files you reviewed
git add charts/numaflow/Chart.yaml charts/numaflow/values.yaml \
        charts/numaflow/crds/ charts/numaflow/templates/ README.md
# Inspect once more before committing.
git diff --cached --stat
git commit -m "chore: v1.8.0 numaflow release"
git push -u origin prep-v1.8.0-release
gh pr create --title "chore: v1.8.0 numaflow release" --base main
```

The commit message format matches the prior releases (`bb33931`, `ce48a7b`, `da10eed`). The PR title should match.

**Before merging, wait for:**
- `lint-test` workflow ✅ — chart-testing lint + KinD install pass.
- A maintainer review ✅.

Do **not** stage `upgrade/.merge-rejects/` — that's a scratch directory and is excluded by convention (consider adding to `.gitignore` if it isn't already).

---

## Step 8 — post-merge verification

Once the PR is merged into `main`, the `release` job in `.github/workflows/ci.yaml` runs automatically:

1. `helm/chart-releaser-action@v1.7.0` packages `charts/numaflow/` into a release artifact.
2. The artifact is published to the `gh-pages` branch and as a GitHub release at <https://github.com/numaproj/helm-charts/releases>.

To verify:

```bash
gh run watch                                                                    # follow the workflow
gh release view numaflow-0.5.0 --repo numaproj/helm-charts                      # confirm release exists
helm repo update numaproj
helm search repo numaproj/numaflow --versions | head -5
```

The new chart version should be searchable within a few minutes of merge.

---

## Troubleshooting

### `sync-mirrored-files` reports `upstream-missing`

The mirrored file's upstream path was renamed, moved, or removed at the new numaflow version. Check the upstream repo for the new location, update `upgrade/internal/mirror/registry.go` accordingly, and re-run the sync.

### `sync-mirrored-files` reports an error like `git is required`

Install `git`, `diff`, and `patch` on the host. All three are pre-flighted by the tool.

### `make upgrade-charts` fails with `Versions are identical`

`Chart.yaml`'s `appVersion` already equals the value you passed in `NUMAFLOW_VERSION`. Either you're trying to re-run a release, or someone bumped `Chart.yaml` out of band. Verify with `cat charts/numaflow/Chart.yaml`.

### A clean merge in `.merged` looks wrong

The tokenizer or the merge logic may have mis-handled a Helm syntax variant. Capture the chart file, the upstream-at-vOld blob, and the upstream-at-vNew blob (the tool prints temp paths in verbose mode); open an issue with all three.

---

## What this playbook deliberately doesn't cover

- **Structural drift detection** (L3 in the design doc) — checks whether the chart has *drifted* from upstream over multiple releases (e.g. a missing block from a release nobody synced). Planned as a follow-up; until then, periodically diff each mirrored file against upstream@<latest> manually.
- **Render-diff sanity check** (L4) — automated comparison of `helm template` output against the prior release's render. Also planned for a follow-up.
