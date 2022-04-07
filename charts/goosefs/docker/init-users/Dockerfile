FROM centos:centos8.2.2004

RUN yum install -y net-tools

COPY *.sh /

RUN chmod +x /*.sh

ENTRYPOINT ["/entrypoint.sh"]
