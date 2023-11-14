#!/usr/bin/env bash
set -x

function printUsage() {
    echo -e "Usage: Run command with related environment variable set"
    echo
    echo -e 'Environment Variable "PORTS_TO_CHECK" is set:'
    echo -e " PORT1:PORT2:PORT3..."
}

function check_port() {
  ports=$1
  for port in "${ports[@]}"; do
    # ignore grep not found
    netstat -ntp | awk '{print $4,"\t",$6,"\t",$7}' | grep "$port"
    if [[ $? -eq 0 ]]; then
      # Found any port is in use
      return 1
    fi
    echo
  done

  # No port in use
  return 0
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

    # Timeout for 60 * 10s(10min)
    count=1
    while [[ count -lt 61 ]]; do
      echo
      echo "Retry to check port usage for the $count time"
      check_port $ports
      if [[ $? == 0 ]]; then
        echo "No port conflict found. Exiting..."
        exit 0
      fi

      count=`expr $count + 1`
      sleep 10
    done

    echo "Timeout for port conflicts"
    exit 1
}

main "$@"
