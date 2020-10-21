# Build the manager binary
FROM golang:1.14.2 as builder

WORKDIR /go/src/github.com/fluid-cloudnative/fluid
COPY . .

# Build
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -gcflags="-N -l" -a -o /go/bin/fluid-controller cmd/controller/main.go

# Debug
#RUN go get github.com/go-delve/delve/cmd/dlv

FROM alpine:3.10
RUN apk add --update curl tzdata iproute2 bash libc6-compat vim &&  \
 	rm -rf /var/cache/apk/* && \
 	cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
 	echo "Asia/Shanghai" >  /etc/timezone

RUN curl -o helm-v3.0.3-linux-amd64.tar.gz http://aliacs-k8s-cn-hangzhou.oss-cn-hangzhou.aliyuncs.com/public/pkg/helm/helm-v3.0.3-linux-amd64.tar.gz && \
    tar -xvf helm-v3.0.3-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/ddc-helm && \
    chmod u+x /usr/local/bin/ddc-helm && \
    rm -f helm-v3.0.3-linux-amd64.tar.gz

ENV K8S_VERSION v1.14.8
RUN curl -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${K8S_VERSION}/bin/linux/amd64/kubectl && chmod +x /usr/local/bin/kubectl

FROM fluid/alpine:helm-kubectl
add charts/ /charts

COPY --from=builder /go/bin/fluid-controller /usr/local/bin/fluid-controller
#COPY --from=builder /go/bin/dlv /usr/local/bin/dlv
RUN chmod -R u+x /usr/local/bin/
CMD fluid-controller

# CMD ["dlv", "--listen=:12345", "exec", "/usr/local/bin/fluid-controller", "--", "--enable-leader-election"]
