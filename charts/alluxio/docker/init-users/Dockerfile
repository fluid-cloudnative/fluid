FROM centos:centos8.2.2004

COPY *.sh /

ADD dataloads/ /dataloads

RUN chmod +x /*.sh && chmod +x /dataloads/*.sh

ENTRYPOINT ["/entrypoint.sh"]
