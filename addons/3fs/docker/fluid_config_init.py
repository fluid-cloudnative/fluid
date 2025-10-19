#!/usr/bin/env python

import json

rawStr = ""
try:
    with open("/etc/fluid/config/config.json", "r") as f:
        rawStr = f.readlines()
except:
    pass

if rawStr == "":
    try:
        with open("/etc/fluid/config.json", "r") as f:
            rawStr = f.readlines()
    except:
        pass

rawStr = rawStr[0]

script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
TOKEN=$(echo $MNT_FROM | awk -F'@' '{print $1}')
RDMA=$(echo $MNT_FROM | awk -F'@' '{print $2}' | awk -F'://' '{print $2}')

echo $TOKEN > /opt/3fs/etc/token.txt

sed -i "s#RDMA://0.0.0.0:8000#${RDMA}#g" /opt/3fs/etc/hf3fs_fuse_main_launcher.toml

MNT_TO=$targetPath
trap "umount ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}

sed -i "s#/3fs/stage#${MNT_TO}#g" /opt/3fs/etc/hf3fs_fuse_main_launcher.toml

LD_LIBRARY_PATH=/opt/3fs/bin /opt/3fs/bin/hf3fs_fuse_main --launcher_cfg /opt/3fs/etc/hf3fs_fuse_main_launcher.toml
"""

obj = json.loads(rawStr)

with open("/mount-3fs.sh", "w") as f:
    f.write('mountPoint="%s"\n' % obj["mounts"][0]["mountPoint"])
    f.write('targetPath="%s"\n' % obj["targetPath"])
    f.write(script)
