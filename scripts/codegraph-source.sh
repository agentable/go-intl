#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'USAGE'
usage: scripts/codegraph-source.sh [init|status|path|clean]

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

copy_source() {
	local root mirror
	root=$(repo_root)
	mirror=$(mirror_dir)

	ensure_safe_mirror
	rm -rf "$mirror"
	mkdir -p "$mirror"

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

			mkdir -p "$mirror/$(dirname "$path")"
			cp -pP "$root/$path" "$mirror/$path"
		done
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
	codegraph status "$(mirror_dir)"
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
