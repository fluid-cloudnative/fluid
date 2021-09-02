FROM centos:centos8.2.2004

COPY *.sh /

RUN yum install -y net-tools

RUN chmod +x /*.sh

ENTRYPOINT ["/entrypoint.sh"]
