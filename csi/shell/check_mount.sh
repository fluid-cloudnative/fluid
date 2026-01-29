#!/bin/bash

set -ex

ConditionPathIsMountPoint="$1"
MountType="$2"
SubPath="$3"

#[ -z ${ConditionPathIsMountPoint} ] && ConditionPathIsMountPoint=/alluxio-fuse

# Retry configuration constants
readonly MAX_RETRY_ATTEMPTS=10
readonly RETRY_SLEEP_INTERVAL=3

# Retry function: retry a command up to MAX_RETRY_ATTEMPTS times with RETRY_SLEEP_INTERVAL
# Avoids eval for security. For commands with pipelines, wrap them in a dedicated function.
# Usage: retry_command <error_message> <exit_code> <command> [args...]
retry_command() {
  local error_message=$1
  local exit_code=$2
  shift 2
  
  local count=0
  while ! "$@" > /dev/null 2>&1
  do
    sleep $RETRY_SLEEP_INTERVAL
    count=$((count + 1))
    if [ $count -eq $MAX_RETRY_ATTEMPTS ]
    then
      echo "$error_message"
      exit $exit_code
    fi
  done
}

# Check if mount point is mounted
check_mount() {
  cat /proc/self/mountinfo | grep -F "$ConditionPathIsMountPoint" | grep -F "$MountType"
}
retry_command "timed out waiting for $ConditionPathIsMountPoint mounted" 1 check_mount

# Check if mount point is accessible
retry_command "timed out stating $ConditionPathIsMountPoint returns ready" 1 \
  stat "$ConditionPathIsMountPoint"

# Check if sub path exists
retry_command "timed out waiting for sub path [$SubPath] to exist" 2 \
  test -e "$ConditionPathIsMountPoint/$SubPath"

echo "succeed in checking mount point $ConditionPathIsMountPoint"
