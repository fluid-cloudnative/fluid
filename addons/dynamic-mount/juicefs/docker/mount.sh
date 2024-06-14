#!/bin/bash

set -e

if [[ $# -ne 4 ]]; then
    echo "Error: require 3 arguments, but got $# arguments"
    exit 1
fi

mount_src=$1        # e.g. juicefs://mybucket
mount_target=$2     # e.g. /runtime-mnt/thin/default/thin-demo/thin-fuse/mybucket
fs_type=$3
mount_opt_file=$4   # e.g. /etc/fluid/mount-opts/mybucket.opts (mount options in json format)

filesystem_name=${mount_src#juicefs://}
token_file=$(cat ${mount_opt_file} | jq -r '.["token"]')
access_key_file=$(cat ${mount_opt_file} | jq -r '.["access-key"]')
secret_key_file=$(cat ${mount_opt_file} | jq -r '.["secret-key"]')
bucket=$(cat ${mount_opt_file} | jq -r '.["bucket"]')

juicefs auth $filesystem_name --token `cat $token_file` --access-key `cat $access_key_file` --secret-key `cat $secret_key_file` --bucket "$bucket"

exec juicefs mount -f $filesystem_name $mount_target