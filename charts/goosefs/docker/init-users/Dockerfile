# FROM centos:centos8.2.2004
FROM centos:centos8.2.2004@sha256:4062bbdd1bb0801b0aa38e0f83dece70fb7a5e9bce223423a68de2d8b784b43b

RUN yum install -y net-tools

COPY *.sh /

RUN chmod +x /*.sh

ENTRYPOINT ["/entrypoint.sh"]
