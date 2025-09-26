#!/bin/bash

set -e

if [[ $# -ne 4 ]]; then
    echo "Error: require 4 arguments, but got $# arguments"
    exit 1
fi

mount_src=$1        # e.g. oss://mybucket/path/to/mydata
mount_target=$2     # e.g. /runtime-mnt/thin/default/thin-demo/thin-fuse/testpath1
fs_type=$3
mount_opt_file=$4   # e.g. /etc/fluid/mount-opts/mybucket.opts (mount options in json format)

# Check if the mount_src starts with oss://
if [[ "$mount_src" == oss://* ]]; then
    oss_full_url=${mount_src#oss://}
else
    echo "Error: mount_src must start with oss://"
    exit 1
fi

bucket_name=$(echo $oss_full_url | cut -d'/' -f1) # extract the bucket name from 
oss_path=${oss_full_url#$bucket_name}
endpoint=$(cat $mount_opt_file | jq -r '.["oss-endpoint"]')
access_key_file=$(cat ${mount_opt_file} | jq -r '.["oss-access-key"]')
secret_key_file=$(cat ${mount_opt_file} | jq -r '.["oss-secret-key"]')

echo "$bucket_name:$(cat $access_key_file):$(cat $secret_key_file)" > /etc/passwd-ossfs
chmod 640 /etc/passwd-ossfs
if [[ -z "$oss_path" ]]; then
    oss_to_mount="$bucket_name:/"
else
    oss_to_mount="$bucket_name:$oss_path"
fi

ossfs -f $oss_to_mount $mount_target -ourl="https://$endpoint"
