# FROM python:3.10
FROM python:3.10@sha256:33f72df2ad8c9f777bf0adb35b9d89c5d62935cee2af1f9c3224fb6f7da1dc6b

# Dependencies are pinned by hash in requirements.txt. Everything comes from
# wheels except lazy-import, which is published as a source distribution only.
COPY requirements.txt /tmp/requirements.txt
RUN pip3 install --no-cache-dir --only-binary=:all: --no-binary=lazy-import --require-hashes \
        -r /tmp/requirements.txt && \
    rm /tmp/requirements.txt

WORKDIR /

ARG APP
ARG APP_DIR
ENV APP=${APP}

COPY ${APP_DIR}/${APP} /${APP}
