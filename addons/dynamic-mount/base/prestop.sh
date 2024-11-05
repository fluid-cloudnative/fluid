#!/bin/bash
set -e

mount_points=$(cat /proc/self/mountinfo | grep " ${MOUNT_POINT}" | awk '{print $5}')

echo "prestop.sh: umounting mountpoints under ${MOUNT_POINT}"
for mount_point in ${mount_points}; do
    echo ">> mount-helper.sh umount ${mount_point}"
    mount-helper.sh umount ${mount_point}
done

# from now on, we clean sub dirs in a best-effort manner.
set +e
echo "prestop.sh: clean sub directories under ${MOUNT_POINT}"
sub_dirs=$(ls "${MOUNT_POINT}/")
for sub_dir in ${sub_dirs}; do
    rmdir "${MOUNT_POINT}/${sub_dir}" || echo "WARNING: failed to rmdir ${sub_dir}, maybe filesystem still mounting on it."
done

exit 0
