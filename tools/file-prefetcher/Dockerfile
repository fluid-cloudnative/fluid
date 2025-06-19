# FROM debian:bookworm-slim
FROM debian:bookworm-slim@sha256:e5865e6858dacc255bead044a7f2d0ad8c362433cfaa5acefb670c1edf54dfef

RUN apt update && apt install -y python3

COPY ./main.py /root/main.py

COPY ./entrypoint.sh /root/entrypoint.sh

RUN chmod u+x /root/entrypoint.sh

ENV PYTHONUNBUFFERED=1

ENTRYPOINT ["/root/entrypoint.sh"]