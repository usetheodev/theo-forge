#!/usr/bin/env bash
# Re-sync testdata/examples/ from upstream argoproj/argo-workflows.
#
# Usage:
#   ./scripts/update-examples.sh v3.6.0
#
# This script is DELIBERATELY interactive: it diffs each upstream YAML
# against ours and prompts before overwriting. The round-trip suite
# (TestRoundTripTestdataExamples) MUST still pass after re-sync, or
# the SDK has drifted away from real Argo semantics.

set -euo pipefail

target_version="${1:-}"
[[ -n "${target_version}" ]] || {
  echo "usage: $0 <argo-tag, e.g. v3.6.0>" >&2
  exit 2
}

repo="https://github.com/argoproj/argo-workflows"
work="$(mktemp -d -t argo-examples-sync.XXXXXX)"
trap 'rm -rf "${work}"' EXIT

echo "==> cloning ${repo} at ${target_version}"
git clone --depth 1 --branch "${target_version}" --filter=blob:none "${repo}" "${work}/repo" >/dev/null

src="${work}/repo/examples"
dst="testdata/examples"
[[ -d "${src}" ]] || { echo "ERROR: ${src} not found in upstream"; exit 1; }

echo "==> diffing"
updated=0
new=0
removed=0

for f in "${src}"/*.yaml; do
  base="$(basename "${f}")"
  ours="${dst}/${base}"
  if [[ ! -f "${ours}" ]]; then
    echo "  NEW upstream: ${base}"
    read -rp "    add to testdata? [y/N] " ans
    [[ "${ans}" == "y" ]] && cp "${f}" "${ours}" && new=$((new+1))
    continue
  fi
  if ! diff -q "${f}" "${ours}" >/dev/null 2>&1; then
    echo "  DRIFTED: ${base}"
    diff -u "${ours}" "${f}" | head -20 || true
    read -rp "    overwrite ours with upstream? [y/N] " ans
    [[ "${ans}" == "y" ]] && cp "${f}" "${ours}" && updated=$((updated+1))
  fi
done

# Removed examples — present locally, gone upstream.
for ours in "${dst}"/*.yaml; do
  base="$(basename "${ours}")"
  if [[ ! -f "${src}/${base}" ]]; then
    echo "  REMOVED upstream: ${base}"
    read -rp "    delete from testdata? [y/N] " ans
    [[ "${ans}" == "y" ]] && rm "${ours}" && removed=$((removed+1))
  fi
done

# Bump VERSION + LAST_SYNCED.
today="$(date +%F)"
sed -i \
  -e "s/^ARGO_VERSION=.*/ARGO_VERSION=${target_version}/" \
  -e "s/^LAST_SYNCED=.*/LAST_SYNCED=${today}/" \
  "${dst}/VERSION"

echo
echo "==> summary"
echo "  updated: ${updated}"
echo "  new:     ${new}"
echo "  removed: ${removed}"
echo "  testdata/examples/VERSION bumped to ${target_version} (${today})"
echo
echo "NEXT: run 'go test -run TestRoundTripTestdataExamples ./...' to confirm semantics."
