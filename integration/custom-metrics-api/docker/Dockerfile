# docker build --no-cache . -t registry.cn-hangzhou.aliyuncs.com/fluid/gen-cert:v0.5.0
FROM alpine:3.10

# Install requirements
ARG KUBE_VERSION="1.18.8"
RUN apk add --update ca-certificates && \
    apk add --update openssl && \
    apk add --update -t deps curl && \
    curl -L https://storage.googleapis.com/kubernetes-release/release/v$KUBE_VERSION/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    apk del --purge deps && \
    rm /var/cache/apk/*

ADD gen-secret.sh /gen-secret.sh

RUN chmod u+x /gen-secret.sh