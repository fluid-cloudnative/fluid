# alpine:3.18
FROM alpine@sha256:11e21d7b981a59554b3f822c49f6e9f57b6068bb74f49c4cd5cc4c663c7e5160 AS builder
ENV OSSFS_VERSION 1.91.1
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories
RUN apk --update add fuse alpine-sdk automake autoconf libxml2-dev fuse-dev curl-dev pkgconf
RUN wget -qO- https://github.com/aliyun/ossfs/archive/v$OSSFS_VERSION.tar.gz |tar xz
RUN cd ossfs-1.91.1 \
  && ./autogen.sh \
  && ./configure --prefix=/usr \
  && make \
  && make install

# alpine:3.18
FROM alpine@sha256:11e21d7b981a59554b3f822c49f6e9f57b6068bb74f49c4cd5cc4c663c7e5160
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories && \
    apk --update add bash coreutils curl fuse inotify-tools libgcc libstdc++ libxml2 openssl python3 tini && \
    rm -rf /var/cache/apk/*
ENV OSSFS_VERSION v1.91.1
COPY --from=builder /usr/bin/ossfs /usr/bin/ossfs

ADD mount_and_umount.py /
ADD inotify.sh /
ADD umount.sh /
RUN chmod u+x /inotify.sh /umount.sh
ENTRYPOINT ["/sbin/tini", "-g", "--"]
