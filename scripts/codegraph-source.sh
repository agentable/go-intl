#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'USAGE'
usage: scripts/codegraph-source.sh [init|status|sync-check|path|clean]

Build and index a source-only CodeGraph mirror for go-intl.

The mirror lives at .tmp/codegraph-source by default and excludes .references,
cache/build directories, and any ignored files. It copies the current worktree,
including tracked modifications and untracked non-ignored files, so CodeGraph
answers describe the code being edited without indexing vendored references.

Environment:
  CODEGRAPH_SOURCE_DIR  Override the mirror directory.
USAGE
}

repo_root() {
	git rev-parse --show-toplevel
}

mirror_dir() {
	local root
	root=$(repo_root)
	printf '%s\n' "${CODEGRAPH_SOURCE_DIR:-$root/.tmp/codegraph-source}"
}

ensure_safe_mirror() {
	local root mirror
	root=$(repo_root)
	mirror=$(mirror_dir)

	case "$mirror" in
	"" | / | "$root" | "$root/" | "$root/.references" | "$root/.references/"*)
		echo "refusing unsafe CODEGRAPH_SOURCE_DIR: $mirror" >&2
		exit 2
		;;
	esac
}

source_paths() {
	local root path
	root=$(repo_root)

	git -C "$root" ls-files -z -co --exclude-standard |
		while IFS= read -r -d '' path; do
			case "$path" in
			.references | .references/* | .git | .git/* | .codegraph | .codegraph/* | .tmp | .tmp/* | bin | bin/* | build | build/*)
				continue
				;;
			esac
			if [[ -d "$root/$path" && ! -L "$root/$path" ]]; then
				continue
			fi
			if [[ ! -e "$root/$path" && ! -L "$root/$path" ]]; then
				continue
			fi
			printf '%s\0' "$path"
		done
}

mirror_paths() {
	local mirror path rel
	mirror=$(mirror_dir)

	[[ -d "$mirror" ]] || return 0
	find "$mirror" \
		\( -path "$mirror/.codegraph" -o -path "$mirror/.codegraph/*" \) -prune -o \
		\( -type f -o -type l \) -print0 |
		while IFS= read -r -d '' path; do
			rel=${path#"$mirror"/}
			printf '%s\0' "$rel"
		done
}

copy_source() {
	local root mirror
	root=$(repo_root)
	mirror=$(mirror_dir)

	ensure_safe_mirror
	rm -rf "$mirror"
	mkdir -p "$mirror"

	source_paths |
		while IFS= read -r -d '' path; do
			mkdir -p "$mirror/$(dirname "$path")"
			cp -pP "$root/$path" "$mirror/$path"
		done
}

sync_check() {
	local root mirror expected actual path mismatch
	root=$(repo_root)
	mirror=$(mirror_dir)
	expected=$(mktemp)
	actual=$(mktemp)
	mismatch=0

	ensure_safe_mirror
	if [[ ! -d "$mirror" ]]; then
		echo "codegraph source mirror missing: $mirror" >&2
		rm -f "$expected" "$actual"
		return 1
	fi

	source_paths | LC_ALL=C sort -z >"$expected"
	mirror_paths | LC_ALL=C sort -z >"$actual"
	if ! cmp -s "$expected" "$actual"; then
		echo "codegraph source mirror file list differs: $mirror" >&2
		diff -u <(tr '\0' '\n' <"$expected") <(tr '\0' '\n' <"$actual") >&2 || true
		rm -f "$expected" "$actual"
		return 1
	fi

	while IFS= read -r -d '' path; do
		if [[ -L "$root/$path" ]]; then
			if [[ ! -L "$mirror/$path" ]]; then
				echo "codegraph source mirror type differs: $path" >&2
				mismatch=1
				continue
			fi
			if [[ "$(readlink "$root/$path")" != "$(readlink "$mirror/$path")" ]]; then
				echo "codegraph source mirror symlink target differs: $path" >&2
				mismatch=1
			fi
			continue
		fi
		if ! cmp -s "$root/$path" "$mirror/$path"; then
			echo "codegraph source mirror content differs: $path" >&2
			mismatch=1
		fi
	done <"$expected"

	rm -f "$expected" "$actual"
	if ((mismatch)); then
		return 1
	fi
	echo "codegraph source mirror is in sync: $mirror"
}

require_codegraph() {
	if ! command -v codegraph >/dev/null 2>&1; then
		echo "codegraph is not installed or not on PATH" >&2
		exit 127
	fi
}

action="${1:-init}"
case "$action" in
init)
	require_codegraph
	copy_source
	codegraph init "$(mirror_dir)"
	;;
status)
	require_codegraph
	sync_check
	codegraph status "$(mirror_dir)"
	;;
sync-check)
	sync_check
	;;
path)
	mirror_dir
	;;
clean)
	ensure_safe_mirror
	rm -rf "$(mirror_dir)"
	;;
-h | --help | help)
	usage
	;;
*)
	usage >&2
	exit 2
	;;
esac
