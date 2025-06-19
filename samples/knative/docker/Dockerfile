# FROM registry.cn-hangzhou.aliyuncs.com/knative-sample/helloworld-go:160e4dc8
FROM registry.cn-hangzhou.aliyuncs.com/knative-sample/helloworld-go:160e4dc8@sha256:032b801f148d959016bd30362c01df96edcaea524182c6b8c07f9ea155f2f44d

RUN apk add bash

COPY entrypoint.sh /

RUN chmod u+x /entrypoint.sh

CMD ["/entrypoint.sh"]
