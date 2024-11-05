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

import json

rawStr = ""
with open("/etc/fluid/config/config.json", "r") as f:
    rawStr = f.readlines()

print(rawStr[0])

script = """
#!/bin/sh
MNT_POINT=$targetPath

echo $MNT_POINT

if test -e ${MNT_POINT}
then
    echo "MNT_POINT exist"
else
    mkdir -p ${MNT_POINT}
fi

/cfs/bin/cfs-client -c /cfs/fuse.json

sleep inf
"""

obj = json.loads(rawStr[0])
volAttrs = obj['mounts'][0]

print("pvAttrs", volAttrs)

fuse = {}
fuse["mountPoint"] = obj["targetPath"]
fuse["volName"] = volAttrs["name"]
fuse["masterAddr"] = volAttrs["mountPoint"]
fuse["owner"] = "root"
fuse["logDir"] = "/cfs/logs/"
fuse["logLevel"] = "error"

print("fuse.json: ", fuse)

with open("/cfs/fuse.json", "w") as f:
    f.write(json.dumps(fuse))

with open("mount-cubefs.sh", "w") as f:
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])
    f.write(script)
