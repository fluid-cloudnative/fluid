FROM debian:bookworm-slim

RUN apt update && apt install -y python3

COPY ./main.py /root/main.py

COPY ./entrypoint.sh /root/entrypoint.sh

RUN chmod u+x /root/entrypoint.sh

ENV PYTHONUNBUFFERED=1

ENTRYPOINT ["/root/entrypoint.sh"]