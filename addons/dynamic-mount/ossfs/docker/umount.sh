#!/bin/bash
set -xe

echo "umount.sh: MOUNT_POINT ${MOUNT_POINT}"
for folder in ${MOUNT_POINT}/*; do
    if [ -d "$folder" ]; then
        echo "umount_all: umount ${folder}"
        umount ${folder}
        rmdir ${folder}
    fi
done