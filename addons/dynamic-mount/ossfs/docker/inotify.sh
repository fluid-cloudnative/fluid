#!/bin/bash
set -xe

MOUNT_INFO="/etc/fluid/config"
mount_script="/etc/fluid/mount"
umount_script="/etc/fluid/umount"

execute_script() {
    script_path="$1"

    # execute all scripts in script_path
    for script in "${script_path}"/*; do
        if [ -f "$script" ]; then
            echo "execute_script: ${script} "
            chmod u+x ${script}
            bash ${script}
        fi
    done
}

mount_and_umount() {
    # delete all mount and umount scripts
    rm -f ${mount_script}/*
    rm -f ${umount_script}/*
    
    mounted=$(cat /proc/self/mountinfo | grep fuse.ossfs | grep ${MOUNT_POINT} | awk '{print $5}')
    
    # fluid_config_parse.py would generate mount and umount scripts if need
    echo "mount_and_umount: python /fluid_config_parse.py ${mounted}"
    python /fluid_config_parse.py ${mounted}

    # execute all scripts
    # 错误处理
    execute_script ${mount_script}
    execute_script ${umount_script}
}

mkdir -p ${mount_script}
mkdir -p ${umount_script}

# mount for init mount info
mount_and_umount

# trap SIGTERM, umount all
trap "/umount.sh" SIGTERM SIGKILL

# if mount info is modified, mount and umount to update
# config.json is mounted by configmap, it is actually a symlink point to actual file,
# and when configmap changes, the old file would be delete and new file would be created,
# so we inotify the delete event
inotifywait -m -r -e delete_self "${MOUNT_INFO}" |
while read -r directory event file; do
    echo "${directory} ${file} changed"
    mount_and_umount
done
