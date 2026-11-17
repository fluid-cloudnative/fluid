set -x -e

# download go package if not exist. The mirror redirects to golang.google.cn
# and then dl.google.com, so redirects are followed but only over HTTPS.
[ -e "go1.13.6.linux-amd64.tar.gz" ] \
  || curl -fsSLO --proto '=https' --tlsv1.2 https://studygolang.com/dl/golang/go1.13.6.linux-amd64.tar.gz

# unpack
[ -d "/usr/local/go" ] || tar -zxf "go1.13.6.linux-amd64.tar.gz" -C "/usr/local"

export PATH=$PATH:/usr/local/go/bin

/alluxio/dev/scripts/generate-tarballs single -target /tmp/alluxio-release-2.8.1-SNAPSHOT-bin.tar.gz
