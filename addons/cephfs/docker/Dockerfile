FROM alpine@sha256:124c7d2707904eea7431fffe91522a01e5a861a624ee31d03372cc1d138a3126
# use alpine:3.17

RUN mkdir /etc/ceph
RUN apk add ceph ceph-fuse python3

ADD fluid_config_init.py /
ADD entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]
