#!/usr/bin/env bash

set -Eeuo pipefail

DEPLOY_HOST="${DEPLOY_HOST:-schedule@43.133.137.16}"
DEPLOY_PATH="${DEPLOY_PATH:-/var/www/jadwallari.com}"
DEPLOY_SERVICE="${DEPLOY_SERVICE:-schedule-core}"
DEPLOY_KEEP_RELEASES="${DEPLOY_KEEP_RELEASES:-5}"
DEPLOY_HEALTHCHECK_URL="${DEPLOY_HEALTHCHECK_URL:-}"

ROOT_DIR="$(git rev-parse --show-toplevel)"
RELEASE_ID="$(date -u +%Y%m%d%H%M%S)"
BUILD_DIR="$(mktemp -d)"
BUILD_BINARY="$BUILD_DIR/schedule-core"

cleanup() {
	rm -rf "$BUILD_DIR"
}
trap cleanup EXIT

die() {
	echo "deploy: $*" >&2
	exit 1
}

[[ "$DEPLOY_KEEP_RELEASES" =~ ^[0-9]+$ ]] || die "DEPLOY_KEEP_RELEASES must be a non-negative integer"

if ! git diff --quiet || ! git diff --cached --quiet; then
	die "working tree is dirty; commit changes before deploying"
fi

echo "Building release $RELEASE_ID..."
(
	cd "$ROOT_DIR"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -trimpath -o "$BUILD_BINARY" ./cmd/api
)

REMOTE_RELEASE_DIR="$DEPLOY_PATH/releases/$RELEASE_ID"
REMOTE_BINARY="$REMOTE_RELEASE_DIR/schedule-core"

ssh_opts=(-o BatchMode=yes -o ConnectTimeout=10)

echo "Preparing $DEPLOY_HOST:$REMOTE_RELEASE_DIR..."
ssh "${ssh_opts[@]}" "$DEPLOY_HOST" bash -s -- "$DEPLOY_PATH" "$RELEASE_ID" <<'REMOTE_SCRIPT'
set -eu

deploy_path="$1"
release_id="$2"

mkdir -p "$deploy_path/releases/$release_id"
test -f "$deploy_path/shared/.env"
REMOTE_SCRIPT

echo "Uploading binary..."
scp "${ssh_opts[@]}" "$BUILD_BINARY" "$DEPLOY_HOST:$REMOTE_BINARY"

echo "Activating release..."
ssh "${ssh_opts[@]}" "$DEPLOY_HOST" bash -s -- \
	"$DEPLOY_PATH" "$RELEASE_ID" "$DEPLOY_SERVICE" "$DEPLOY_KEEP_RELEASES" <<'REMOTE_SCRIPT'
set -eu

deploy_path="$1"
release_id="$2"
service="$3"
keep_releases="$4"
release_dir="$deploy_path/releases/$release_id"

chmod 755 "$release_dir/schedule-core"
ln -sfn "$deploy_path/shared/.env" "$release_dir/.env"
rm -f "$deploy_path/current.next"
ln -s "$release_dir" "$deploy_path/current.next"
mv -Tf "$deploy_path/current.next" "$deploy_path/current"

sudo -n /bin/systemctl restart "$service"
for attempt in 1 2 3 4 5 6 7 8 9 10; do
	if systemctl is-active --quiet "$service"; then
		break
	fi
	sleep 1
done
systemctl is-active --quiet "$service"

active_release="$(readlink -f "$deploy_path/current")"
find "$deploy_path/releases" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' \
	| sort -r \
	| tail -n "+$((keep_releases + 1))" \
	| while read -r old_release; do
		old_path="$deploy_path/releases/$old_release"
		if [[ "$old_path" != "$active_release" ]]; then
			rm -rf -- "$old_path"
		fi
	done
REMOTE_SCRIPT

if [[ -n "$DEPLOY_HEALTHCHECK_URL" ]]; then
	echo "Running health check..."
	curl --fail --silent --show-error "$DEPLOY_HEALTHCHECK_URL" >/dev/null
fi

echo "Deployed $RELEASE_ID to $DEPLOY_HOST:$DEPLOY_PATH/current"
