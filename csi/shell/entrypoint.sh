#!/usr/bin/env bash
set -xe

rm -f /var/lib/kubelet/csi-plugins/fuse.csi.fluid.io、csi.sock
mkdir -p /var/lib/kubelet/csi-plugins/fuse.csi.fluid.io

fluid-csi $@