#!/bin/bash

set -ex

ConditionPathIsMountPoint="$1"
MountType="$2"
SubPath="$3"

#[ -z ${ConditionPathIsMountPoint} ] && ConditionPathIsMountPoint=/alluxio-fuse

count=0
# while ! mount | grep alluxio | grep  $ConditionPathIsMountPoint | grep -v grep
while ! cat /proc/self/mountinfo | grep $ConditionPathIsMountPoint | grep $MountType
do
    sleep 3
    count=`expr $count + 1`
    if test $count -eq 10
    then
        echo "timed out waiting for $ConditionPathIsMountPoint mounted"
        exit 1
    fi
done

count=0
while ! stat $ConditionPathIsMountPoint
do
  sleep 3
  count=`expr $count + 1`
  if test $count -eq 10
    then
        echo "timed out stating $ConditionPathIsMountPoint returns ready"
        exit 1
    fi 
done

if [ ! -e  $ConditionPathIsMountPoint/$SubPath ] ; then
  echo "sub path [$SubPath] not exist!"
  exit 2
fi

echo "succeed in checking mount point $ConditionPathIsMountPoint"
