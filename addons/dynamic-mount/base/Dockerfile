FROM debian:bullseye@sha256:a165446a88794db4fec31e35e9441433f9552ae048fb1ed26df352d2b537cb96 as builder

RUN apt update && \
    apt install -y --no-install-recommends build-essential ca-certificates git libfuse3-dev pkg-config python3-pip

# meson and ninja come from wheels only, pinned by hash in requirements.txt.
COPY requirements.txt /tmp/requirements.txt
RUN pip install --no-cache-dir --only-binary=:all: --require-hashes -r /tmp/requirements.txt


RUN git clone https://github.com/libfuse/libfuse.git && cd libfuse && git checkout fuse-3.16.2 && cd ..

RUN mkdir -p libfuse/build && cd libfuse/build && meson setup .. && ninja install

RUN cd libfuse/example && gcc -Wall passthrough.c `pkg-config fuse3 --cflags --libs` -o passthrough

FROM debian:bullseye-slim@sha256:a165446a88794db4fec31e35e9441433f9552ae048fb1ed26df352d2b537cb96

# ca-certificates is only a Recommends of python3, but mount programs built on
# this image need it to reach HTTPS endpoints.
RUN apt update && \
    apt install -y --no-install-recommends ca-certificates fuse inotify-tools jq python3 supervisor tini && \
    rm -rf /var/cache/apt/* && \
    ln -s /usr/bin/python3 /usr/local/bin/python
COPY inotify-fluid-config.ini /tmp/inotify-fluid-config.ini
RUN cat /tmp/inotify-fluid-config.ini >> /etc/supervisor/supervisord.conf && rm /tmp/inotify-fluid-config.ini

COPY reconcile_mount_program_settings.py mount-helper.sh inotify.sh mount-passthrough-fuse.sh prestop.sh entrypoint.sh /usr/local/bin/
RUN chmod u+x /usr/local/bin/mount-helper.sh /usr/local/bin/inotify.sh /usr/local/bin/mount-passthrough-fuse.sh /usr/local/bin/prestop.sh /usr/local/bin/entrypoint.sh

RUN apt update && apt install -y --no-install-recommends libfuse3-3 fuse3
COPY --from=builder libfuse/example/passthrough /usr/local/bin/passthrough

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
