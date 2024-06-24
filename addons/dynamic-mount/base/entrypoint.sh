#!/bin/bash

set -e

trap "/usr/local/bin/prestop.sh" SIGTERM 

if [[ "$USE_PASSTHROUGH_FUSE" == "True" ]]; then
mkdir -p $MOUNT_POINT
cat << EOF >> /etc/supervisor/supervisord.conf

[program:passthrough-fuse]
command=/usr/local/bin/mount-passthrough-fuse.sh
redirect_stderr=true
stdout_logfile=/proc/1/fd/1
stdout_logfile_maxbytes=0
autorestart=true
startretries=9999
EOF
fi

supervisord -n
