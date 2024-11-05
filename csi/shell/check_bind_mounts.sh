#!/bin/bash

set -ex

VolumeName="$1"

MountPattern="pods/.*/volumes/kubernetes\.io~csi/$VolumeName/mount"

cat /proc/self/mountinfo | grep $MountPattern
