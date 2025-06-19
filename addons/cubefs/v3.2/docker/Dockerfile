#FROM chubaofs/cfs-client:v3.2.0
FROM chubaofs/cfs-client:v3.2.0@sha256:8254359f256164c67825bb6a035ce77baa773ee0775f7346f29fffabea6b97d2

COPY fluid_config_init.py /
COPY entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]