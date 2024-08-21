# shellcheck shell=bash

# Run commands with elevated privileges.

# Check if the current user is root
function perm::check_if_root() {
  if [[ $EUID -ne 0 ]]; then
    return 1
  else
    return 0
  fi
}

# Prefix commands with sudo or not
function perm::run_as_root() {
  if [[ $EUID -ne 0 ]]; then
    sudo "$@"
  else
    "$@"
  fi
}
