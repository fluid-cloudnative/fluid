#!/usr/bin/env bash
set -e

function printUsage() {
    echo -e "Usage: Run command with related environment variable set"
    echo
    echo -e 'Environment Variable "PORTS_TO_CHECK" is set:'
    echo -e " PORT1:PORT2:PORT3..."
}

function main() {
    # The shell scripts only reports the usage status of the ports.
    # If any port is in use, no err will be returned.
    ports="$PORTS_TO_CHECK"
    ports=(${ports//:/ })
    if [[ "${#ports[*]}" -eq 0 ]]; then
        printUsage
        exit 1
    fi
    for port in "${ports[@]}"; do
      echo "Checking if port $port is in use:"
      echo "> netstat -nltp | grep $port"
      # ignore grep not found
      netstat -nltp | grep "$port" || [[ $? == 1 ]]
      echo
    done

}

main "$@"