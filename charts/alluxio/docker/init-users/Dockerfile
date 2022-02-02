FROM centos:centos8.2.2004

COPY *.sh /

RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-Linux-*

RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-Linux-*

RUN yum install -y net-tools

RUN chmod +x /*.sh

ENTRYPOINT ["/entrypoint.sh"]
