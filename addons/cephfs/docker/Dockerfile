FROM alpine

RUN mkdir /etc/ceph
RUN apk add ceph ceph-fuse python3

ADD fluid_config_init.py /
ADD entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]
