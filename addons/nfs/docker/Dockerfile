# Build environment
FROM ubuntu:jammy as BUILD
RUN apt update && \
    apt install --yes automake libfuse-dev libnfs-dev libnfs-dev libnfs13 libtool libtool m4 make xsltproc


COPY ./fuse-nfs-master /src
WORKDIR /src
RUN ./setup.sh && \
    ./configure && \
    make

# Production image
FROM ubuntu:jammy
RUN apt update && \
    apt install --yes bash fuse libfuse2 libnfs13 python3 && \
    apt clean autoclean && \
    apt autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
ADD ./fluid_config_init.py /
ADD ./entrypoint.sh /usr/local/bin
COPY --from=BUILD /src/fuse/fuse-nfs /bin/fuse-nfs
CMD ["/usr/local/bin/entrypoint.sh"]