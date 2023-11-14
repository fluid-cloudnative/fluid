FROM chubaofs/cfs-client:2.4.0

ADD fluid_config_init.py /
ADD entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]