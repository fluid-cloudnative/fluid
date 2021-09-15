#!/bin/bash

set -ex

VolumeName="$1"

MountPattern="pods/.*/volumes/kubernetes\.io~csi/$VolumeName/mount"

mount | grep $MountPattern
