# Build environment
FROM ubuntu:jammy as BUILD
RUN apt update && \
    apt install --yes libfuse-dev libnfs13 libnfs-dev libtool m4 automake libnfs-dev xsltproc make libtool


COPY ./fuse-nfs-master /src
WORKDIR /src
RUN ./setup.sh && \
    ./configure && \
    make

# Production image
FROM ubuntu:jammy
RUN apt update && \
    apt install --yes libnfs13 libfuse2 fuse python3 bash && \
    apt clean autoclean && \
    apt autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
ADD ./fluid_config_init.py /
ADD ./entrypoint.sh /usr/local/bin
COPY --from=BUILD /src/fuse/fuse-nfs /bin/fuse-nfs
CMD ["/usr/local/bin/entrypoint.sh"]