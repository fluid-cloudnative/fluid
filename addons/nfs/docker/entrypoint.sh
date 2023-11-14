#!/usr/bin/env bash
set +x

python3 /fluid_config_init.py

chmod u+x /mount-nfs.sh

bash /mount-nfs.sh