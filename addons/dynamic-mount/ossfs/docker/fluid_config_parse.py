#  Copyright 2023 The Fluid Authors.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

#!/usr/bin/env python

import json,os,sys

rawStr = ""
with open("/etc/fluid/config/config.json", "r") as f:
    rawStr = f.readlines()

print(rawStr[0])

mount_script = """
#!/bin/sh
set -e

akId=`cat $accessKey`
akSecret=`cat $accessSecret`
echo "${bucket}:${akId}:${akSecret}" > /etc/passwd-ossfs
chmod 640 /etc/passwd-ossfs

if test -d ${MNT_POINT}
then
    echo "MNT_POINT exist"
else
    mkdir -p ${MNT_POINT}
fi

echo "mount command: ossfs ${BUCKET_PATH} ${MNT_POINT} ${OSSFS_OPTIONS}"
ossfs ${BUCKET_PATH} ${MNT_POINT} ${OSSFS_OPTIONS}
"""

umount_script="""
#!/bin/sh
set -e

if ! test -d ${MNT_POINT}
then
    echo "${MNT_POINT} does not exist"
    exit
fi

if [ $(mount | grep fuse | grep -c "${MNT_POINT}") -ne 0 ]; then
    umount ${MNT_POINT}
    rm -rf ${MNT_POINT}
    echo "umount ${MNT_POINT} successfully."
else
    echo "${MNT_POINT} is not a fuse mountpoint."
fi
"""

obj = json.loads(rawStr[0])

if len(sys.argv) > 1:
    mounted_path = sys.argv[1:]
else:
    mounted_path = []
mounted = []
for p in mounted_path:
    if not p.startswith("/"):
        continue
    mounted.append(p.split("/")[-1])
target_mounted = [mount["name"] for mount in obj["mounts"]]

need_mount = list(set(target_mounted).difference(set(mounted)))
need_unmount = list(set(mounted).difference(set(target_mounted)))

mount_folder = "/etc/fluid/mount"

for mount in obj["mounts"]:
    if mount["name"] not in need_mount:
        continue
    bucket = mount["mountPoint"].lstrip("oss://")
    bucketPath = bucket
    path = mount.get("path")
    if path is not None:
        bucketPath += ":/"+path if not path.startswith("/") else ":"+path
    options = "-ourl={}".format(mount["options"]["url"])
    # parse more options here
    if mount["options"].get("allow_other") is not None:
        options += " -oallow_other"
    targetPath = os.path.join(obj["targetPath"], mount["name"])
    mount_script_path = "{}/mount-{}.sh".format(mount_folder, mount["name"])
    with open(mount_script_path, "w") as f:
        f.write("bucket=\"%s\"\n" % bucket)
        f.write("accessKey=\"%s\"\n" % mount["options"]["oss-access-key"])
        f.write("accessSecret=\"%s\"\n" % mount["options"]["oss-access-secret"])
        f.write("BUCKET_PATH=\"%s\"\n" % bucketPath)
        f.write("MNT_POINT=\"%s\"\n" % targetPath)
        f.write("OSSFS_OPTIONS=\"%s\"\n" % options)
        f.write(mount_script)

umount_folder="/etc/fluid/umount"

for name in need_unmount:
    umount_script_path = "{}/umount-{}.sh".format(umount_folder, name)
    with open(umount_script_path, "w") as f:
        targetPath = os.path.join(obj["targetPath"], name)
        f.write("MNT_POINT=\"%s\"\n" % targetPath)
        f.write(umount_script)
