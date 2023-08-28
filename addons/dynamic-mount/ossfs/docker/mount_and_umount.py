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
import shutil
import subprocess
import time

PASSWD="/etc/passwd-ossfs"

def check_bucket_in_passwd(bucket):
    if not os.path.exists(PASSWD):
        return False
    with open(PASSWD, 'r') as passwd_file:
        for line in passwd_file:
            if bucket in line:
                return True
    return False

def secret(bucket, access_key, access_secret):
    if check_bucket_in_passwd(bucket):
        return
    # get access_key and access_secret from file
    with open(access_key, 'r') as f:
        access_key = f.read().strip()
    with open(access_secret, 'r') as f:
        access_secret = f.read().strip()

    credentials = f"{bucket}:{access_key}:{access_secret}\n"

    # write ossfs credentials to /etc/passwd-ossfs
    with open(PASSWD, 'a') as f:
        f.write(credentials)
    os.chmod(PASSWD, 640)

def is_fuse_mount(mnt_point):
    cmd = f"cat /proc/self/mountinfo | grep fuse | grep {mnt_point}"
    proc = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = proc.communicate()
    if error:
        raise Exception(f"Failed to execute command: {cmd}, error message: {error}")
    return bool(output)
    

def mount_ossfs(bucket_path, mnt_point, ossfs_options):
    if not os.path.exists(mnt_point):
        # create dirs is not exist
        os.makedirs(mnt_point)
    if is_fuse_mount(mnt_point):
        # if mnt_point is fuse mountpoint, pass mount
        print(f"{mnt_point} is a fuse mount, skip mount")
        return
    retry_count = 3
    while retry_count:
        try:
            print("mount_ossfs: ossfs", bucket_path, mnt_point, ossfs_options)
            subprocess.check_call(['ossfs', bucket_path, mnt_point] + ossfs_options.split())
            return
        except subprocess.CalledProcessError as e:
            print(f"Failed to mount, error: {e}")
            retry_count -= 1
            if retry_count:
                print("retry...")
                time.sleep(1) 
            else:
                raise 

def umount_ossfs(mnt_point):
    if not os.path.exists(mnt_point):
        # pass not exist mnt_point
        print(f"{mnt_point} does not exist")
        return

    if not is_fuse_mount(mnt_point):
        # if mnt_point is not fuse mountpoint, remove
        print(f"{mnt_point} is not a FUSE mount point")
        try:
            shutil.rmtree(mnt_point)
        except OSError as e:
            print(f"Failed to remove {mnt_point}, error: {e}")
        return

    try:
        print("umount_ossfs: umount", mnt_point)
        subprocess.check_call(['umount', mnt_point])
    except subprocess.CalledProcessError as e:
        print(f"Failed to unmount {mnt_point}, error: {e}")
        return
    
    try:
        shutil.rmtree(mnt_point)
    except OSError as e:
        print(f"Failed to remove {mnt_point}, error: {e}")

if __name__=="__main__":
    rawStr = ""
    with open("/etc/fluid/config/config.json", "r") as f:
        rawStr = f.readlines()

    print(rawStr[0])

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
    print(f"need mount: {need_mount}, need umount: {need_unmount}")

    for mount in obj["mounts"]:
        if mount["name"] not in need_mount:
            continue
        bucket = mount["mountPoint"].lstrip("oss://")
        bucket_path = bucket
        path = mount.get("path")
        if path is not None:
            bucket_path += ":/"+path if not path.startswith("/") else ":"+path
        options = "-ourl={}".format(mount["options"]["url"])
        # parse more options here
        if mount["options"].get("allow_other") is not None:
            options += " -oallow_other"
        target_path = os.path.join(obj["targetPath"], mount["name"])
        secret(bucket, mount["options"]["oss-access-key"], mount["options"]["oss-access-secret"])
        mount_ossfs(bucket_path, target_path, options)

    for name in need_unmount:
        target_path = os.path.join(obj["targetPath"], name)
        umount_ossfs(target_path)
