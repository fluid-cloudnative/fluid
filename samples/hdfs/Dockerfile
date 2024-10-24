FROM adoptopenjdk/openjdk11:jre-11.0.8_10-centos
MAINTAINER "yukong.lxx@alibaba-inc.com"

WORKDIR /workspace

COPY target/fluid-hdfs-demo.jar /workspace/

ENTRYPOINT [ "java", "-jar", "fluid-hdfs-demo.jar" ]