#!/bin/bash

# This script is used to automatically generate tarball,
# build and upload docker images to AliYun
# on given branch,tag or commit.

set -e -u -o pipefail

# Following arguments are initialized with the default value.
#alluxio_git='https://github.com/Alluxio/alluxio.git'
alluxio_git='https://github.com/Alluxio/alluxio.git'
branch="release-2.8.0"
tag=""
commit=""
alluxio_image_name="fluidcloudnative/alluxio"
alluxio_fuse_image_name="fluidcloudnative/alluxio-fuse"

dev_container_name="alluxio-dev-test"

print_usage()
{
  echo -e "Usage:"
  echo -e "\t./build-image.sh [options]"
  echo -e "OPTIONS:"
  echo -e "\t-h"
  echo -e "\t\tDisplay this help message."
  echo -e "\t-b, --branch branch"
  echo -e "\t\tSet the git branch."
  echo -e "\t-t, --tag tag"
  echo -e "\t\tSet the git tag."
  echo -e "\t-c, --commit commmit_id"
  echo -e "\t\tSet the commit id."
  echo -e "\t-a, --alluxio_image_name alluxio_image_name"
  echo -e "\t\tSet the alluxio image name."
  echo -e "\t-f, --alluxio_fuse_image_name alluxio_fuse_image_name"
  echo -e "\t\tSet the alluxio fuse image name."
}

clean()
{
  docker rm -f ${dev_container_name} || true
  # rm -rf /alluxio || true
}

clone()
{
  if [ -d "/alluxio" ]; then
    echo "alluxio repository already exists."
  else
    echo "cloning from ${alluxio_git}."
    git clone "${alluxio_git}" /alluxio
  fi
}

checkout()
{
  local owd=$(pwd)
  # Make sure that local repository is up tp date
  cd "/alluxio" && git fetch
  # 1. checkout commit
  if [[ -n ${commit} ]]; then
    echo "checkout to commit ${commit}."
    git checkout ${commit}
  elif [[ -n ${tag} ]]; then
  # 2. checkout tag
    echo "checkout to tag ${tag}."
    git checkout ${tag}
  # 3. checkout branch
  else
    echo "checkout to branch ${branch}."
    git checkout "remotes/origin/${branch}"
  fi
  echo "GIT_COMMIT=$(git rev-parse --short HEAD)"
  cd "${owd}"
}

start_container()
{
  docker pull "maven:3.6.2-jdk-8"
  local dev_container_id=$(docker ps | grep ${dev_container_name} | awk '{print $1}')
  if [ -z ${dev_container_id} ]; then
    echo "start maven container."
    docker run -itd -v /alluxio:/alluxio --name "${dev_container_name}" "maven:3.6.2-jdk-8" bash
    dev_container_id=$(docker ps | grep ${dev_container_name} | awk '{print $1}')
  fi

  if [ -z ${dev_container_id} ]; then
    echo "ERROR: can not start container." >&2
    exit 1
  fi
}

tarball()
{
  docker cp tarball.sh ${dev_container_name}:/tarball.sh
  docker exec -it ${dev_container_name} /bin/bash -c "sh /tarball.sh"
}

build()
{
  docker cp ${dev_container_name}:/tmp/alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz /tmp/
  cp /tmp/alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz /alluxio/integration/docker

  cd /alluxio/integration/docker

  GIT_COMMIT=$(git rev-parse --short HEAD)
  echo "GIT_COMMIT=${GIT_COMMIT}"

  # Use aluxio/alluxio-dev image for both Alluxio and Fuse, starting from Alluxio v2.6.2
  # Build Alluxio image
  if test -f /alluxio/integration/docker/Dockerfile-dev; then
    docker build -f Dockerfile-dev -t alluxio/alluxio-dev:release-2.8.1-SNAPSHOT-$GIT_COMMIT --build-arg ALLUXIO_TARBALL=alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz --build-arg ENABLE_DYNAMIC_USER="true" .
  else
    docker build -t alluxio/alluxio-dev:release-2.8.1-SNAPSHOT-$GIT_COMMIT --build-arg ALLUXIO_TARBALL=alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz --build-arg ENABLE_DYNAMIC_USER="true" .
  fi
  # Tag Alluxio image
  docker tag alluxio/alluxio-dev:release-2.8.1-SNAPSHOT-$GIT_COMMIT ${alluxio_image_name}:release-2.8.1-SNAPSHOT-$GIT_COMMIT

  # Build Fuse image if needed. Tag Fuse image
  if test -f /alluxio/integration/docker/Dockerfile.fuse; then
    docker build -f Dockerfile.fuse -t alluxio/alluxio-fuse:release-2.8.1-SNAPSHOT-$GIT_COMMIT --build-arg ALLUXIO_TARBALL=alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz --build-arg ENABLE_DYNAMIC_USER="true" .
    docker tag alluxio/alluxio-fuse:release-2.8.1-SNAPSHOT-$GIT_COMMIT ${alluxio_fuse_image_name}:release-2.8.1-SNAPSHOT-$GIT_COMMIT
  else
    docker tag alluxio/alluxio-dev:release-2.8.1-SNAPSHOT-$GIT_COMMIT ${alluxio_fuse_image_name}:release-2.8.1-SNAPSHOT-$GIT_COMMIT
  fi

  docker push ${alluxio_fuse_image_name}:release-2.8.1-SNAPSHOT-$GIT_COMMIT &
  docker push ${alluxio_image_name}:release-2.8.1-SNAPSHOT-$GIT_COMMIT &
}

main()
{
  # Parse arguments using getopt
  ARGS=$(getopt -a -o b:c:t:a:f:h --long branch:,commit:,tag:,alluxio_image_name:,alluxio_fuse_image_name:,help -- "$@")
  if [ $? != 0 ]; then
    exit 1
  fi

  eval set -- "${ARGS}"

  while true
  do
    case "$1" in
      -h|--help)
        print_usage
        shift 1
        exit 0
        ;;
      -b|--branch)
        branch=$2
        echo "branch=$2"
        shift 2
        ;;
      -c|--commit)
        commit=$2
        echo "commit=$2"
        shift 2
        ;;
      -t|--tag)
        tag=$2
        echo "tag=$2"
        shift 2
        ;;
      -a|--alluxio_image_name)
        alluxio_image_name=$2
        echo "alluxio_image_name=$2"
        shift 2
        ;;
      -f|--alluxio_fuse_image_name)
        alluxio_fuse_image_name=$2
        echo "alluxio_fuse_image_name=$2"
        shift 2
        ;;
      --)
        shift
        break
        ;;
      *)
        echo "ERROR: invalide argument $1" >&2
        exit 1
        ;;
    esac
  done

  clean && clone && checkout && start_container && tarball && build

  if [ $? == 0 ]; then
    echo "Build SUCCESS!"
  else
    echo "Build FAILED!"
    exit 1
  fi
}

main "$@"
