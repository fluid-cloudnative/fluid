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
with open("/etc/fluid/config.json", "r") as f:
    rawStr = f.readlines()

rawStr = rawStr[0]

script = """
#!/bin/sh
set -ex
MNT_FROM=$mountPoint
MNT_TO=$targetPath


trap "fusermount -u ${MNT_TO}" SIGTERM
mkdir -p ${MNT_TO}
fuse-nfs -n nfs://${MNT_FROM} -m ${MNT_TO}
sleep inf
"""

obj = json.loads(rawStr)

with open("mount-nfs.sh", "w") as f:
    f.write("mountPoint=\"%s\"\n" % obj['mounts'][0]['mountPoint'])
    f.write("targetPath=\"%s\"\n" % obj['targetPath'])
    f.write(script)