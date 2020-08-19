FROM python:2.7.17-alpine3.10

RUN echo "http://dl-8.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

# Add HDF5 support
RUN apk add --no-cache --allow-untrusted --repository http://dl-3.alpinelinux.org/alpine/edge/testing hdf5 hdf5-dev
RUN apk --no-cache --update-cache add gcc gfortran python python-dev py-pip build-base wget freetype-dev libpng-dev openblas-dev bash
RUN ln -s /usr/include/locale.h /usr/include/xlocale.h

RUN pip install numpy six

RUN mkdir /app

ADD . /app

RUN chmod u+x /app/*.sh

# ENV COCO 10

ENV THREADS 1

WORKDIR /app
