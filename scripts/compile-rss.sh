#!/usr/bin/env bash
#
# compile-rss.sh — evidence collector for issue #3 (CLDR compile-memory blowup).
#
# WHAT THIS IS
#   This is an EVIDENCE COLLECTOR, not a gate. It reproduces issue #3 — the
#   `internal/cldr` package compiling at multi-GB resident memory and OOM-killing
#   downstream CI — and prints the peak RSS of the Go `compile` subprocess for each
#   target package, alongside wall-clock time.
#
# WHY IT IS NOT A HARD GATE
#   The numbers here are observations, not invariants. Peak compile RSS drifts with
#   the Go compiler version, host architecture, and SSA heuristics — a threshold on
#   such a number would be a budget, and a budget invites growth. The CI gates test
#   the CAUSE instead (payload literal shape + import-graph boundaries); this
#   script stays documentation-only and is never wired into CI.
#   Run it manually to record before/after RSS in a PR as evidence.
#
# USAGE
#   scripts/compile-rss.sh [package ...]
#   With no arguments it measures each leaf domain package under internal/cldr
#   (the retired root internal/cldr bomb package no longer exists).
#
# METHOD / TRADEOFF
#   To force a real recompile we point GOCACHE at a fresh temporary directory per
#   package and build with `-p 1`. Tradeoff vs. `go build -a`:
#     - A temporary GOCACHE isolates each package's measurement (no cross-run cache
#       bleed) and is trivially restored by deleting the temp dir, leaving the user's
#       real build cache untouched.
#     - `-p 1` serializes the toolchain so the sampled peak reflects the single
#       heaviest `compile` invocation (the target package), matching how issue #3
#       OOMs on one subprocess, rather than the sum of parallel compiles.
#     - Both this and `go build -a` recompile dependencies too; that is acceptable
#       because the target package dominates and we report its peak.
#   We sample RSS by polling `ps` for live `compile` processes (the Go toolchain's
#   compiler). macOS has no /proc, so /usr/bin/time-style RSS readout is unportable;
#   polling `ps -o rss=` works identically on darwin and linux, so the same script
#   reproduces the number on a CI runner.
#
#   To identify the compiler we match on `ps -o args=` (the full command line,
#   a.k.a. `command`), NOT `ps -o comm=`. `comm=` is unportable for this purpose:
#   BSD/darwin prints the full executable path, but Linux truncates it to the
#   command name capped at 15 chars (TASK_COMM_LEN) — i.e. literally `compile` and
#   never the `${GOTOOLDIR}/compile` absolute path. The first whitespace-delimited
#   token of `args=` is the absolute executable path on both OSes, so matching it
#   against `${GOTOOLDIR}/compile` works identically on darwin and linux.

set -euo pipefail

# Resolve the Go compiler binary path so we sample only its processes.
GOTOOLDIR="$(go env GOTOOLDIR)"
COMPILE_BIN="${GOTOOLDIR}/compile"

# Polling interval in seconds. Fine-grained enough to catch the peak of a
# multi-second compile without burning CPU on the sampler itself.
SAMPLE_INTERVAL="0.05"

# dir_has_go_files reports whether a directory contains at least one .go file,
# without relying on bash-4 globbing builtins (macOS ships bash 3.2).
dir_has_go_files() {
  local f
  for f in "$1"*.go; do
    [ -e "$f" ] && return 0
  done
  return 1
}

# Default targets: the monolithic root package plus each leaf domain.
default_targets() {
  # The root internal/cldr package — the original bomb — is retired: every domain
  # owns its data directly and internal/cldr is no longer a Go package. Measure
  # each surviving domain package instead.
  local dir
  for dir in internal/cldr/*/; do
    [ -d "$dir" ] || continue
    # Only directories that actually hold a Go package.
    if dir_has_go_files "$dir"; then
      printf './%s\n' "${dir%/}"
    fi
  done
}

# peak_rss_kb prints the maximum RSS (in KB) seen across all live `compile`
# processes whose executable matches COMPILE_BIN. Portable between darwin and linux.
sample_compile_rss_kb() {
  # `ps -axo rss=,args=` lists every process with its RSS and full command line.
  # The first whitespace-delimited token of the command line ($2) is the absolute
  # executable path on both darwin and linux, so we keep rows whose executable is
  # exactly the Go compiler, then take the max. We deliberately avoid `comm=`,
  # which Linux truncates to a 15-char command name (never the absolute path).
  ps -axo rss=,args= 2>/dev/null \
    | awk -v bin="$COMPILE_BIN" '
        $2 == bin { if ($1 > max) max = $1 }
        END { print max + 0 }
      '
}

# build_one runs an isolated recompile of one package while polling compile RSS.
# It echoes: "<peak_rss_kb> <wall_seconds>".
build_one() {
  local pkg="$1"
  local tmp_cache
  tmp_cache="$(mktemp -d "${TMPDIR:-/tmp}/compile-rss.XXXXXX")"

  # Whole-second resolution: macOS date(1) has no %N, and compile times of
  # interest here are tens of seconds.
  local start_s end_s
  start_s="$(date +%s)"

  # Launch the build in the background with an isolated cache to force a full
  # recompile of the target package and its dependencies.
  GOCACHE="$tmp_cache" go build -p 1 "$pkg" >/dev/null 2>&1 &
  local build_pid=$!

  local peak_kb=0 cur_kb
  while kill -0 "$build_pid" 2>/dev/null; do
    cur_kb="$(sample_compile_rss_kb)"
    if [ "$cur_kb" -gt "$peak_kb" ]; then
      peak_kb="$cur_kb"
    fi
    sleep "$SAMPLE_INTERVAL"
  done

  # Surface a failed build instead of silently reporting a tiny RSS.
  if ! wait "$build_pid"; then
    rm -rf "$tmp_cache"
    echo "build failed for ${pkg}" >&2
    return 1
  fi

  end_s="$(date +%s)"
  rm -rf "$tmp_cache"

  printf '%s %s\n' "$peak_kb" "$((end_s - start_s))"
}

main() {
  local targets=()
  if [ "$#" -gt 0 ]; then
    targets=("$@")
  else
    # Read default targets line by line; portable across bash 3.2 (no mapfile).
    local line
    while IFS= read -r line; do
      [ -n "$line" ] && targets+=("$line")
    done < <(default_targets)
  fi

  # Column header.
  printf '%-40s %14s %12s\n' "PACKAGE" "PEAK RSS (MB)" "WALL (s)"
  printf '%-40s %14s %12s\n' "----------------------------------------" "--------------" "------------"

  local pkg result peak_kb wall_s peak_mb
  for pkg in "${targets[@]}"; do
    result="$(build_one "$pkg")"
    peak_kb="${result%% *}"
    wall_s="${result##* }"
    # KB -> MB, integer rounding.
    peak_mb="$(( (peak_kb + 512) / 1024 ))"
    printf '%-40s %14s %12s\n' "$pkg" "$peak_mb" "$wall_s"
  done
}

main "$@"
