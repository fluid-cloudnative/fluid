import json
import os
import re
import subprocess

obj = json.load(open("/etc/fluid/config/config.json"))

mount_point = obj["mounts"][0]["mountPoint"]
target_path = obj["targetPath"]

# Normalize first to resolve redundant separators, '.' and '..' components
target_path = os.path.normpath(target_path)

# Validate that the normalized path is an absolute POSIX path
if not os.path.isabs(target_path) or not target_path.startswith('/'):
    print(f"Error: target_path must be absolute: {target_path}")
    exit(1)

# Safety check: ensure no '..' components remain after normalization
if '..' in target_path.split('/'):
    print(f"Error: Path traversal using '..' is not allowed in target_path: {target_path}")
    exit(1)

# Validate that the path contains only safe characters
if not re.match(r'^[/a-zA-Z0-9._-]+$', target_path):
    print(f"Error: target_path contains invalid characters: {target_path}")
    exit(1)

# Prevent mounting on the root directory
if target_path == '/':
    print("Error: target_path resolves to the root directory '/' and is not allowed.")
    exit(1)

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
