FROM alpine:3.18 AS builder
ENV OSSFS_VERSION 1.91.1
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories
RUN apk --update add fuse alpine-sdk automake autoconf libxml2-dev fuse-dev curl-dev pkgconf
RUN wget -qO- https://github.com/aliyun/ossfs/archive/v$OSSFS_VERSION.tar.gz |tar xz
RUN cd ossfs-1.91.1 \
  && ./autogen.sh \
  && ./configure --prefix=/usr \
  && make \
  && make install

FROM alpine:3.18
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories
RUN apk --update add fuse curl libxml2 openssl libstdc++ libgcc python3 inotify-tools bash coreutils tini && rm -rf /var/cache/apk/* 
ENV OSSFS_VERSION v1.91.1
COPY --from=builder /usr/bin/ossfs /usr/bin/ossfs

ADD gen_mount_script.py /
ADD inotify.sh /
ADD umount.sh /
RUN chmod u+x /inotify.sh /umount.sh
ENTRYPOINT ["/sbin/tini", "-g", "--"]