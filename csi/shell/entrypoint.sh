#!/usr/bin/env bash
set -xe

# Function to check if a directory exists
check_directory() {
  local dir="$1"
  
  if [ ! -d "$dir" ]; then
    echo "Error: $dir does not exist."
    echo "See https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/troubleshooting/debug-fuse.md#%E6%AD%A5%E9%AA%A41-1 for more information!"
    exit 1
  fi
}

check_directory "$KUBELET_ROOTDIR/pods"
check_directory "$KUBELET_ROOTDIR/plugins"

rm -f $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io/csi.sock
mkdir -p $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io

fluid-csi start $@
