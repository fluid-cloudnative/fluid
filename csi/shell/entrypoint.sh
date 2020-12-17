#!/usr/bin/env bash
set -xe

rm -f $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io/csi.sock
mkdir -p $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io

fluid-csi start $@
