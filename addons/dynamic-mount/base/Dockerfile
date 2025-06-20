FROM debian:bullseye@sha256:a165446a88794db4fec31e35e9441433f9552ae048fb1ed26df352d2b537cb96 as builder

RUN apt update && \
    apt install -y build-essential git libfuse3-dev pkg-config python3-pip

# RUN pip install meson ninja
# RUN pip install meson==1.8.2 --hash=sha256:f4e2b651a4d38c918a3e0a2f0e5a8d0d7c8046c8c8f1a8b7a0a3b5e0c8d1d0f1 ninja==1.11.1.4  --hash=sha256:5713cf50c5be50084a8693308a63ecf9e55c3132a78a41ab1363a28b6caaaee1

RUN git clone https://github.com/libfuse/libfuse.git && cd libfuse && git checkout fuse-3.16.2 && cd ..

RUN mkdir -p libfuse/build && cd libfuse/build && meson setup .. && ninja install

RUN cd libfuse/example && gcc -Wall passthrough.c `pkg-config fuse3 --cflags --libs` -o passthrough

FROM debian:bullseye-slim@sha256:a165446a88794db4fec31e35e9441433f9552ae048fb1ed26df352d2b537cb96

RUN apt update && \
    apt install -y fuse inotify-tools jq python3 supervisor tini && \
    rm -rf /var/cache/apt/* && \
    ln -s /usr/bin/python3 /usr/local/bin/python
COPY inotify-fluid-config.ini /tmp/inotify-fluid-config.ini
RUN cat /tmp/inotify-fluid-config.ini >> /etc/supervisor/supervisord.conf && rm /tmp/inotify-fluid-config.ini

COPY reconcile_mount_program_settings.py mount-helper.sh inotify.sh mount-passthrough-fuse.sh prestop.sh entrypoint.sh /usr/local/bin/
RUN chmod u+x /usr/local/bin/mount-helper.sh /usr/local/bin/inotify.sh /usr/local/bin/mount-passthrough-fuse.sh /usr/local/bin/prestop.sh /usr/local/bin/entrypoint.sh

RUN apt update && apt install -y libfuse3-3 fuse3
COPY --from=builder libfuse/example/passthrough /usr/local/bin/passthrough

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
