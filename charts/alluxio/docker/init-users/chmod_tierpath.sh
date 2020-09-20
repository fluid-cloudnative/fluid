#!/usr/bin/env bash
set -e

function printUsage() {
  echo -e "Usage: Run command with related environment variable set"
  echo
  echo -e 'Environment Variable "$FLUID_TIERSTORE_PATHS" is set:'
  echo -e " PATH1:PATH2:PATH3..."

}

function main() {
#    if [[ -z "$FLUID_TIERSTORE_PATH" ]]; then
#        echo -e "Env variable FLUID_TIERSTORE_PATH is empty"
#        exit 1
#    fi
    paths="$FLUID_TIERSTORE_PATHS"
    paths=(${paths//:/ })
    if [[ "${#paths[*]}" -eq 0 ]]; then
      printUsage
      exit 1
    fi
    for path in "${paths[@]}"; do
        chmod -R 0777 $path
    done
}

main "$@"

