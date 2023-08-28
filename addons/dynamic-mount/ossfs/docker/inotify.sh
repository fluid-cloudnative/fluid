#!/bin/bash
set -xe

FUSE_CONFIG="/etc/fluid/config"

mount_and_umount() {
    mounted=$(cat /proc/self/mountinfo | grep fuse.ossfs | grep ${MOUNT_POINT} | awk '{print $5}')
    python /mount_and_umount.py ${mounted}
}

# mount for init fuse-config
mount_and_umount

# trap SIGTERM, umount all
trap "/umount.sh" SIGTERM SIGKILL

# if fuse-config is modified, mount and umount to update
# config.json is mounted by configmap, it is actually a symlink point to actual file,
# and kubernetes would replace actual file by create new one and delete old one, so we handle inotify deletion event.
# Please see https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/util/atomic_writer.go#L93-L138 for more information
inotifywait -m -r -e delete_self "${FUSE_CONFIG}" |
    while read -r directory event file; do
        echo "${directory} ${file} changed"
        mount_and_umount
    done
