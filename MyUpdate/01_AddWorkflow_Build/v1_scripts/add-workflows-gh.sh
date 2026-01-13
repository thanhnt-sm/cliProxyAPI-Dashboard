#!/usr/bin/env bash
set -eu

# Usage: ./add-workflows-gh.sh <owner/repo> <upstream-owner/upstream-repo> [branch]
REPO="${1:-thanhnt-sm/cliProxyAPI-Dashboard}"
UPSTREAM="${2:-original-owner/original-repo}"
BRANCH="${3:-main}"

TMPDIR="$(mktemp -d)"
echo "Using tmp dir: $TMPDIR"

git clone "https://github.com/${REPO}.git" "$TMPDIR/repo"
cd "$TMPDIR/repo"
git checkout "$BRANCH" || git checkout -b "$BRANCH"

mkdir -p .github/workflows

# Write release.yml
cat > .github/workflows/release.yml <<'YML'
# (paste the full release.yml content here exactly as in the provided block)
YML

# Write upstream-check.yml and replace UPSTREAM_REPO placeholder
cat > .github/workflows/upstream-check.yml <<'YML'
# (paste the full upstream-check.yml content here exactly as in the provided block)
YML
# Replace placeholder upstream with provided upstream
sed -i.bak "s|original-owner/original-repo|${UPSTREAM}|g" .github/workflows/upstream-check.yml || true
rm -f .github/workflows/upstream-check.yml.bak

git add .github/workflows/release.yml .github/workflows/upstream-check.yml
git commit -m "Add release + upstream-check workflows"
git push origin "$BRANCH"

echo "Workflows pushed to ${REPO} on branch ${BRANCH}."
echo "Cleanup tmp dir."
rm -rf "$TMPDIR"