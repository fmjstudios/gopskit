#!/usr/bin/env bash
#
# Manage the package release workflow for 'gopskit': building all applications, creating a combined tarball of all
# applications and their LICENSE, as well as a separate tarball for each application within a predefined build
# directory.
#

set -o errexit
#set -o nounset
set -o pipefail

# Libraries
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)

# shellcheck source=scripts/lib/paths.sh
. "${SCRIPT_DIR}/lib/paths.sh"

# shellcheck source=scripts/lib/stdout.sh
. "${SCRIPT_DIR}/lib/stdout.sh"

# Constants
BUILD_DIR="$(lib::paths::root)/dist"
BUILD_TMP_DIR="${BUILD_DIR}/tmp"

PLATFORM="linux"
BAZEL_CACHE_PATH="${HOME}/.cache/bazel"

# File references
license="$(lib::paths::root)/LICENSE"

# query package names and labels from Bazel
PKGS=()
LABELS=()

# ----------------------
#   'create_directories' function
# ----------------------
function create_directories() {
  if [[ ! -d "$BUILD_DIR" ]]; then
    log::yellow "Creating build directory: $BUILD_DIR"
    mkdir -p "$BUILD_DIR"
  else
    log::red "Removing old build directory: $BUILD_DIR"
    rm -rf "$BUILD_DIR"
  fi

  # intentional
  if [[ ! -d "$BUILD_TMP_DIR" ]]; then
    log::yellow "Creating temporary directory within build directory: $BUILD_TMP_DIR"
    mkdir -p "$BUILD_TMP_DIR"
  fi

  log::green "Created build directories!"
  return 0
}

# ----------------------
#   'build_executables' function
# ----------------------
function build_executables() {
  bazel --output_user_root="$BAZEL_CACHE_PATH" build //... 2>/dev/null
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Bazel build failed. Could not build executables!"
    return 1
  fi

  log::green "Built 'gopskit' executables!"
  return 0
}

# ----------------------
#   'build_tarballs' function
# ----------------------
function build_tarballs() {
  for lbl in "${LABELS[@]}"; do
    pkg=$(echo "$lbl" | cut -d':' -f 2)
    log::yellow "Copying $pkg binary to $BUILD_TMP_DIR"

    output=$(bazel --output_user_root="$BAZEL_CACHE_PATH" cquery --output=files "$lbl" 2>/dev/null)
    destination=$BUILD_TMP_DIR
    cp "$output" "$destination"
  done

  log::yellow "Copying package LICENSE to $BUILD_TMP_DIR"
  cp "$license" "$BUILD_TMP_DIR"

  for pkg in "${PKGS[@]}"; do
    log::yellow "Building tarball for $pkg"
    output=$(printf "%s/%s_%s.tar.gz" "$BUILD_DIR" "$pkg" "$PLATFORM")

    tar czfvP "$output" -C "$BUILD_TMP_DIR" "$pkg" LICENSE >/dev/null
    rc=$?
    if [[ $rc -ne 0 ]]; then
      log::red "Could not build tarball for package: $pkg"
      return "$rc"
    fi
  done

  log::yellow "Building 'gopskit' tarball"
  tarb=$(printf "%s/gopskit_%s.tar.gz" "$BUILD_DIR" "$PLATFORM")
  tar czfvP "$tarb" -C "$BUILD_TMP_DIR" "${PKGS[@]}" "LICENSE" >/dev/null
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Could not build 'gopskit' tarball"
    return "$rc"
  fi

  log::green "Built 'gopskit' executables!"
  return 0
}

# ----------------------
#   'calculate_checksums' function
# ----------------------
function calculate_checksums() {
  log::yellow "Changing working directory to $BUILD_DIR"
  old_pwd=$(pwd)
  cd "$BUILD_DIR"

  for pkg in "${PKGS[@]}"; do
    log::yellow "Calculating checksums for $pkg"
    output=$(printf "%s_%s.tar.gz" "$pkg" "$PLATFORM")
    sha256sum "$output" >>"$BUILD_DIR/checksums.txt"
    rc=$?
    if [[ $rc -ne 0 ]]; then
      log::red "Could not calculate SHA256 checksums for package: $pkg"
      return "$rc"
    fi
  done

  log::yellow "Calculating checksums for 'gopskit'"
  tarba=$(printf "gopskit_%s.tar.gz" "$PLATFORM")
  sha256sum "$tarba" >>"$BUILD_DIR/checksums.txt"
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Could not calculate SHA256 checksums for 'gopskit' tarball"
    return "$rc"
  fi

  # switch back
  cd "$old_pwd"

  log::green "Calculated 'gopskit' SHA256 checksums!"
  return 0
}

# --------------------------------
#   MAIN
# --------------------------------
function main() {
  # initialize constants
  raw=$(bazel --output_user_root="$BAZEL_CACHE_PATH" query 'kind("go_binary", //...)' 2>/dev/null)
  while IFS="$(printf '\n')" read -r line; do LABELS+=("$line"); done <<<"$raw"
  while IFS="$(printf '\n')" read -r line; do PKGS+=("$(echo "$line" | cut -d':' -f 2)"); done <<<"$raw"
  unset IFS

  # CI config
  if [[ "$CI" == "true" ]]; then
    if [[ "$OS" =~ "ubuntu" ]]; then
      PLATFORM="linux"
    elif [[ "$OS" =~ "macos" ]]; then
      PLATFORM="darwin"
    elif [[ "$OS" =~ "windows" ]]; then
      PLATFORM="windows"
    fi

    if [[ -n "$CACHE_PATH" ]]; then
      BAZEL_CACHE_PATH="$CACHE_PATH"
    fi
  fi

  # create build directories
  create_directories
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Could not create build directories!"
    return "$rc"
  fi

  # build Go binaries
  build_executables
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Could not build 'gopskit' executables!"
    return "$rc"
  fi

  # build tarballs
  build_tarballs
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Could not build 'gopskit' tarballs!"
    return "$rc"
  fi

  # checksums
  calculate_checksums
  rc=$?
  if [[ $rc -ne 0 ]]; then
    log::red "Could not calculate 'gopskit' checksums!"
    return "$rc"
  fi

  # cleanup
  if [[ -d "$BUILD_TMP_DIR" ]]; then
    log::yellow "Removing temporary directory within build directory: $BUILD_TMP_DIR"
    rm -rf "$BUILD_TMP_DIR"
  fi

  log::green "release.sh finished!"
}

# ------------
# 'main' call
# ------------
main "$@"
