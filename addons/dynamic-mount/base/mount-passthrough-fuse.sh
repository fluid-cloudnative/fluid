#!/bin/bash
set -ex

umount $MOUNT_POINT || true
passthrough -o modules=subdir,subdir=/mnt,auto_unmount -f $MOUNT_POINT
