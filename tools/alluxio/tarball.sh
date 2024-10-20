set -x -e

# download go package if not exist
[ -e "go1.13.6.linux-amd64.tar.gz" ] \
  || wget https://studygolang.com/dl/golang/go1.13.6.linux-amd64.tar.gz

# unpack
[ -d "/usr/local/go" ] || tar -zxf "go1.13.6.linux-amd64.tar.gz" -C "/usr/local"

export PATH=$PATH:/usr/local/go/bin

/alluxio/dev/scripts/generate-tarballs single -target /tmp/alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz
