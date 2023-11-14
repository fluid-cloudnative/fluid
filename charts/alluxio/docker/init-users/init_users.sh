#!/usr/bin/env bash
set -xe

function printUsage() {
    echo -e "Usage: Run command with related environment variable set"
    echo
    echo -e 'Environment Variable "$FLUID_INIT_USERS" is set:'
    echo -e " UID:UserName:GID,GroupID1:GroupName1..."
}

function main() {
    args="$FLUID_INIT_USERS"
    args=(${args//,/ })
    if [[ "${#args[*]}" -lt 2 ]]; then
        printUsage
        exit 1
    fi

    user=${args[0]}
    user_kv=(${user//:/ })
    uid=${user_kv[0]}
    username=${user_kv[1]}
    gid=${user_kv[2]}

    # create groups
    $(>temp)
    echo -n "useradd -m -u ${uid} -g ${gid} -G 0," >>temp
    for ((num = 1; num < ${#args[*]}; num++)); do
        group="${args[${num}]}"
        group_kv=(${group//:/ })
        groupid=${group_kv[0]}
        groupname=${group_kv[1]}
        echo -n "${groupid}" >>temp
        if [[ num -ne $((${#args[*]} - 1)) ]]; then
            echo -n "," >>temp
        fi
        groupadd -f -g ${groupid} ${groupname}
    done

    # create user and bind to group
    echo -n " ${username}" >>temp
    temp=$(cat temp)
    groups=${temp}
    eval $groups
    cat /etc/passwd >/tmp/passwd
    cat /etc/group >/tmp/group
}

main "$@"
