#  Copyright 2022 The Fluid Authors.
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

import json


def write_conf(pvAttrs: dict):
    confAttrs = pvAttrs
    with open("/etc/ceph/ceph.conf", "w") as f:
        f.write("[global]\n")
        f.write("fsid=%s\n" % confAttrs["fsid"])
        f.write("mon_initial_members=%s\n" % confAttrs["mon_initial_members"])
        f.write("mon_host=%s\n" % confAttrs["mon_host"])
        f.write("auth_cluster_required=%s\n" % confAttrs["auth_cluster_required"])
        f.write("auth_service_required=%s\n" % confAttrs["auth_service_required"])
        f.write("auth_client_required=%s\n" % confAttrs["auth_client_required"])


def write_keyring(pvAttrs: dict):
    keyringAttrs = pvAttrs
    with open("/etc/ceph/ceph.client.admin.keyring", "w+") as f:
        f.write("[client.admin]\n")
        f.write("key=%s\n" % keyringAttrs["key"])


def read_json():
    with open("/etc/fluid/config/config.json", "r") as f:
        rawStr = f.readlines()
    rawStr = "".join(rawStr)
    obj = json.loads(rawStr)
    return obj


def write_cmd(mon_url: str, target_path):
    mon_url = mon_url.replace("ceph://", "")
    script = """#!/bin/sh
mkdir -p {}
exec ceph-fuse -n client.admin -k /etc/ceph/ceph.client.admin.keyring -c /etc/ceph/ceph.conf  {}
sleep inf
"""
    with open("/mount_ceph.sh", "w+") as f:
        f.write(script.format(target_path, target_path))


if __name__ == '__main__':
    pvAttrs = read_json()
    write_conf(pvAttrs['mounts'][0]['options'])
    write_keyring(pvAttrs['mounts'][0]['options'])
    write_cmd(pvAttrs['mounts'][0]['mountPoint'], pvAttrs['targetPath'])
    