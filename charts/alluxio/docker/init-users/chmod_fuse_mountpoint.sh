#!/usr/bin/env bash
set -e

function printUsage() {
    echo -e "Usage: Run command with related environment variable set"
    echo
    echo -e 'Environment Variable "$FLUID_FUSE_MOUNTPOINT" is set:'
    echo -e " PATH1:PATH2:PATH3..."

}

function main() {
    if [[ -z "$FLUID_FUSE_MOUNTPOINT" ]]; then
        printUsage
        exit 1
    fi
    chmod -R 0777 $FLUID_FUSE_MOUNTPOINT
}

main "$@"
