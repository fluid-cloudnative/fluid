#!/usr/bin/env sh
set +x

python3 /fluid_config_init.py

chmod u+x /mount_ceph.sh

sh /mount_ceph.sh
