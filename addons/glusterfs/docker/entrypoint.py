import json
import os
import subprocess

obj = json.load(open("/etc/fluid/config/config.json"))

mount_point = obj["mounts"][0]["mountPoint"]
target_path = obj["targetPath"]

os.makedirs(target_path, exist_ok=True)

if len(mount_point.split(":")) != 2:
    print(
        f"The mountPoint format [{mount_point}] is wrong, should be server:volumeId")
    exit(1)

server, volume_id = mount_point.split(":")
args = ["glusterfs", "--volfile-server", server, "--volfile-id",
        volume_id, target_path, "--no-daemon", "--log-file", "/dev/stdout"]

# Available options are described in the following pages:
# https://manpages.ubuntu.com/manpages/trusty/en/man8/mount.glusterfs.8.html
# https://manpages.ubuntu.com/manpages/trusty/en/man8/glusterfs.8.html
if "options" in obj["mounts"][0]:
    options = obj["mounts"][0]["options"]
    for option in options:
        if option[0] == "ro":
            option[0] = "read-only"
        elif option[0] == "transport":
            option[0] = "volfile-server-transport" 
            
        if option[1].lower() == "true":
            args.append(f'--{option[0]}')
        elif option[1].lower() == "false":
            continue
        else:
            args.append(f"--{option[0]}={option[1]}")

subprocess.run(args)
