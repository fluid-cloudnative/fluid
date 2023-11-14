#!/usr/bin/env bash
set -e

function printUsage() {
    echo "Usage: COMMAND [COMMAND_OPTIONS]"
    echo
    echo "COMMAND is one of:"
    echo -e " init_users"
    echo -e " chmod_tierpath"
    echo -e " chmod_fuse_mountpoint"
    echo -e " check_port"
}

function main() {
    if [[ "$#" -eq 0 ]]; then
        printUsage
        exit 1
    fi
    while [[ ! "$#" -eq 0 ]]; do
        case "${1}" in
        init_users)
            sh -c ./init_users.sh
            ;;
        chmod_tierpath)
            sh -c ./chmod_tierpath.sh
            ;;
        chmod_fuse_mountpoint)
            sh -c ./chmod_fuse_mountpoint.sh
            ;;
        check_port)
            sh -c ./check_port.sh
            ;;
        *)
            printUsage
            ;;
        esac
        shift
    done
}

main "$@"
