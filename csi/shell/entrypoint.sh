#!/usr/bin/env bash
set -xe

# Check if required subfolders exist in $KUBELET_ROOTDIR folder
if [ ! -d "$KUBELET_ROOTDIR/pods" ]; then
    echo "Error: $KUBELET_ROOTDIR does not contain /pods folder, it is not a kubelet rootdir."
    echo "See https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/troubleshooting/debug-fuse.md#%E6%AD%A5%E9%AA%A41-1 for more information!"
    exit 1
fi

if [ ! -d "$KUBELET_ROOTDIR/plugins" ]; then
    echo "Error: $KUBELET_ROOTDIR does not contain /plugins folder, it is not a kubelet rootdir."
    echo "See https://github.com/fluid-cloudnative/fluid/blob/master/docs/zh/troubleshooting/debug-fuse.md#%E6%AD%A5%E9%AA%A41-1 for more information!"
    exit 1
fi

rm -f $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io/csi.sock
mkdir -p $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io

fluid-csi start $@
