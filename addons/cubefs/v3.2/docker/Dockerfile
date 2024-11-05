FROM chubaofs/cfs-client:v3.2.0

COPY fluid_config_init.py /
COPY entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]