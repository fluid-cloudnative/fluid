#!/usr/bin/env bash
set -xe

# if KUBELET_ROOTDIR not contains config.yaml, it is not likely a real kubelet rootdir, exit with 1
if [ ! "$(ls $KUBELET_ROOTDIR/config.yaml)" ] ; then
  echo "KUBELET_ROOTDIR [$KUBELET_ROOTDIR] is not likely a real kubelet rootdir, please configure csi.kubelet.rootDir in helm charts! "
  exit 1
fi

rm -f $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io/csi.sock
mkdir -p $KUBELET_ROOTDIR/csi-plugins/fuse.csi.fluid.io

fluid-csi start $@
