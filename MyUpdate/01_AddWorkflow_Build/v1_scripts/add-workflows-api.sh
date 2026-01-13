#!/usr/bin/env bash
set -eu

# Usage: GITHUB_TOKEN=<your_PAT> ./add-workflows-api.sh <owner/repo> <upstream-owner/upstream-repo> [branch]
REPO="${1:-thanhnt-sm/cliProxyAPI-Dashboard}"
UPSTREAM="${2:-original-owner/original-repo}"
BRANCH="${3:-main}"

if [ -z "${GITHUB_TOKEN:-}" ]; then
  echo "Please export GITHUB_TOKEN (a PAT with repo scope) before running"
  exit 1
fi

API="https://api.github.com/repos/${REPO}/contents"
COMMIT_MSG="Add release + upstream-check workflows"
TMPDIR="$(mktemp -d)"
echo "Using tmp dir $TMPDIR"

# release.yml content
cat > "$TMPDIR/release.yml" <<'YML'
# (paste the full release.yml content here exactly as in the provided block)
YML

# upstream-check.yml content (placeholder will be replaced)
cat > "$TMPDIR/upstream-check.yml" <<'YML'
# (paste the full upstream-check.yml content here exactly as in the provided block)
YML

# replace placeholder upstream in file
sed -i.bak "s|original-owner/original-repo|${UPSTREAM}|g" "$TMPDIR/upstream-check.yml" || true
rm -f "$TMPDIR/upstream-check.yml.bak"

upload_file() {
  local path="$1"
  local file="$2"
  local url="${API}/${path}"
  local content
  content=$(base64 -w 0 < "$file")

  resp=$(curl -sS -H "Authorization: Bearer $GITHUB_TOKEN" -H "Accept: application/vnd.github+json" "$url")
  sha=$(echo "$resp" | jq -r '.sha // empty')

  if [ -n "$sha" ]; then
    echo "Updating $path (sha: $sha)"
    body=$(jq -n --arg m "$COMMIT_MSG" --arg c "$content" --arg b "$BRANCH" --arg s "$sha" '{message:$m,content:$c,branch:$b,sha:$s}')
  else
    echo "Creating $path"
    body=$(jq -n --arg m "$COMMIT_MSG" --arg c "$content" --arg b "$BRANCH" '{message:$m,content:$c,branch:$b}')
  fi

  curl -sS -X PUT -H "Authorization: Bearer $GITHUB_TOKEN" -H "Accept: application/vnd.github+json" \
    -d "$body" "$url" | jq -r '.content.path, .commit.sha'
}

upload_file ".github/workflows/release.yml" "$TMPDIR/release.yml"
upload_file ".github/workflows/upstream-check.yml" "$TMPDIR/upstream-check.yml"

echo "Done. Clean up $TMPDIR"
rm -rf "$TMPDIR"