# FROM fluidcloudnative/fluid-dynamic-mount-base:v0.4
FROM fluidcloudnative/fluid-dynamic-mount-base:v0.4@sha256:b4c037b23d96d40ee67f1a578e6193c369fdbb12c38134725ca353babe3a30cd
WORKDIR /root
RUN apt update && \
    apt-get install -y automake autotools-dev g++ git libcurl4-gnutls-dev libfuse-dev libssl-dev libxml2-dev make pkg-config && \
    git clone -b v1.91.8 https://github.com/aliyun/ossfs.git && \
    cd ossfs && \
    ./autogen.sh && \
    ./configure && \
    make && \
    make install && \
    cd .. && \
    rm -rf ossfs

RUN ossfs --version
RUN apt install -y jq 
COPY mount.sh /opt/mount.sh
RUN chmod u+x /opt/mount.sh
