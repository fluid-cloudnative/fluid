#!/usr/bin/env bash
set -e

function printUsage {
  echo "Usage: COMMAND [COMMAND_OPTIONS]"
  echo
  echo "COMMAND_OPTION is one of:"
  echo -e " UID:UserName:GID GroupID:GroupName..."
}

function main {
    if [[ "$#" -lt 2 ]]; then
      printUsage
      exit 1
    fi
    
    user="$1"
    user_kv=(${user//:/ })
    uid=${user_kv[0]}
    username=${user_kv[1]}
    gid=${user_kv[2]}

    groupadd -f -g ${gid} ${username}

    # create groups
    $(> temp)
    echo -n "useradd -m -u ${uid} -g ${gid} -G " >> temp 
    for ((num=2; num<=$#; num++)) ; do
        group="${!num}"
        group_kv=(${group//:/ })
        groupid=${group_kv[0]}
        groupname=${group_kv[1]}
        echo -n "${groupid}" >> temp 
        if [[ num -ne $# ]]; then
            echo -n "," >> temp 
        fi
        groupadd -f -g ${groupid} ${groupname}
    done

    # create user and bind to group
    echo -n " ${username}" >> temp 
    temp=$(cat temp) 
    groups=${temp}
    eval $groups
}

main "$@"
