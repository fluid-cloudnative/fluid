FROM registry.cn-hangzhou.aliyuncs.com/knative-sample/helloworld-go:160e4dc8

RUN apk add bash

ADD entrypoint.sh /

RUN chmod u+x /entrypoint.sh

CMD ["/entrypoint.sh"]
