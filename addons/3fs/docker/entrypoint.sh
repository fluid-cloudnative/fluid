#!/usr/bin/env bash
set +x

echo "sleep inf" > /mount-3fs.sh
python3 /fluid_config_init.py
chmod u+x /mount-3fs.sh
bash /mount-3fs.sh
