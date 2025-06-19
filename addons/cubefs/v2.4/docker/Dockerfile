FROM chubaofs/cfs-client:2.4.0@sha256:2035c53400c4b065e8706f0cc48660ab1c9be342560600abb0e6a69dbee3eb41
# FROM chubaofs/cfs-client:2.4.0

COPY fluid_config_init.py /
COPY entrypoint.sh /usr/local/bin

RUN chmod u+x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]