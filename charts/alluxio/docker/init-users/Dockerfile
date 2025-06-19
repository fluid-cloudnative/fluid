# FROM centos:centos8.2.2004
FROM centos:centos8.2.2004@sha256:4062bbdd1bb0801b0aa38e0f83dece70fb7a5e9bce223423a68de2d8b784b43b

COPY *.sh /

RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*.repo && \
sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*.repo && \
yum install -y net-tools

RUN chmod +x /*.sh

ENTRYPOINT ["/entrypoint.sh"]
