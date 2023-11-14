#!/usr/bin/env bash
set +x

python /fluid_config_init.py

chmod u+x /mount-cubefs.sh

bash /mount-cubefs.sh