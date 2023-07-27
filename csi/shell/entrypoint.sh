#!/usr/bin/env bash
set -xe

# Function to check if a directory exists
check_kubelet_rootdir_subfolder() {
  local dir="$1"
  
  if [ ! -d "$dir" ]; then
    echo "Error: subfolder $dir does not exist, please check whether KUBELET_ROOTDIR $KUBELET_ROOTDIR is configured correctly." 
    echo "Please see https://github.com/fluid-cloudnative/fluid/blob/master/docs/en/userguide/install.md#advanced-configuration for more information!"
    exit 1
  fi
}

check_kubelet_rootdir_subfolder "$KUBELET_ROOTDIR/pods"
check_kubelet_rootdir_subfolder "$KUBELET_ROOTDIR/plugins"

rm -f "$KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io/csi.sock"
mkdir -p "$KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io"

fluid-csi start $@
