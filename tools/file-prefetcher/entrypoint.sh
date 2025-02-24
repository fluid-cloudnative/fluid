#!/bin/bash

if [[ ! -e "/tmp/fluid-file-prefetcher/status/prefetcher.status" ]]; then
    python3 /root/main.py
fi

exec sleep inf