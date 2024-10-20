#!/bin/bash
set -xe

FUSE_CONFIG="/etc/fluid/config"

python /usr/local/bin/reconcile_mount_program_settings.py
supervisorctl update

# if fuse-config(/etc/fluid/config/config.json) is modified, reconcile setting files under /etc/supervisor.d and use `supervisorctl update` to start/stop new/old fuse daemon process.
# config.json is mounted by configmap, it is actually a symlink point to actual file, and kubernetes would atomically rename ..data_tmp to ..data, which triggers an inotify moved_to event.
# Please see https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/util/atomic_writer.go#L93-L138 for more information
inotifywait -m -r -e moved_to "${FUSE_CONFIG}" |
    while read -r directory event file; do
        echo "${directory} ${file} changed (event: ${event})"
        # mount_and_umount
        python /usr/local/bin/reconcile_mount_program_settings.py 
        supervisorctl update
    done
