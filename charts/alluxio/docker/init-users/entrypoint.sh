#!/usr/bin/env bash
set -e

function printUsage() {
   echo -e "Usage: sss"
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
            *)
                printUsage
                ;;
        esac
        shift
    done
}

main "$@"
