FROM shaowenchen/demo-3fsbuilder:latest
RUN apt-get install -y python3
COPY bin /opt/3fs/bin
RUN chmod +x /opt/3fs/bin/*  && mkdir -p /var/log/3fs
COPY etc /opt/3fs/etc
COPY ./fluid_config_init.py /
COPY ./entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh
ENTRYPOINT []
