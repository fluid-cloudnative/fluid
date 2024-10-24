#!/bin/bash

set -e

if [[ $# -ne 4 ]]; then
    echo "Error: require 3 arguments, but got $# arguments"
    exit 1
fi

mount_src=$1        # e.g. juicefs://volume-name/subpath
mount_target=$2     # e.g. /runtime-mnt/thin/default/thin-demo/thin-fuse/mybucket
fs_type=$3
mount_opt_file=$4   # e.g. /etc/fluid/mount-opts/mybucket.opts (mount options in json format)

# Check if the mount_src starts with juicefs://
if [[ "$mount_src" == juicefs://* ]]; then
    juicefs_volume_path=${mount_src#juicefs://}
else
    echo "Error: mount_src does not start with juicefs://"
    exit 1
fi

if [[ "$juicefs_volume_path" == *"/"* ]]; then
    volume_name="${juicefs_volume_path%%/*}"
    volume_subpath="${juicefs_volume_path#*/}"
else
    volume_name=$juicefs_volume_path
fi

token_file=$(cat ${mount_opt_file} | jq -r '.["token"]')
access_key_file=$(cat ${mount_opt_file} | jq -r '.["access-key"]')
secret_key_file=$(cat ${mount_opt_file} | jq -r '.["secret-key"]')
bucket=$(cat ${mount_opt_file} | jq -r '.["bucket"]')

juicefs auth $volume_name --token `cat $token_file` --access-key `cat $access_key_file` --secret-key `cat $secret_key_file` --bucket "$bucket"

if [[ -z "$volume_subpath" ]]; then
# e.g. juicefs://volume-name/ || juicefs://volume-name
    exec juicefs mount -f $volume_name $mount_target
else
# e.g. juicefs://volume-name/subpath-1 || juicefs://volume-name/subpath-1/subpath-2/.../
    exec juicefs mount -f $volume_name $mount_target --subdir="/"$volume_subpath
fi
