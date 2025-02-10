FROM debian:bookworm-slim

RUN apt update && apt install -y python3

ADD ./main.py /root/main.py

ADD ./entrypoint.sh /root/entrypoint.sh

RUN chmod u+x /root/entrypoint.sh

ENV PYTHONUNBUFFERED=1

ENTRYPOINT ["/root/entrypoint.sh"]