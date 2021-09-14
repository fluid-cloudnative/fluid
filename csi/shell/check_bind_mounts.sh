#!/bin/bash

set -ex

VolumeName="$1"

MountPattern="/var/lib/kubelet/pods/.*/volumes/kubernetes\.io~csi/$VolumeName/mount"

mount | grep $MountPattern
