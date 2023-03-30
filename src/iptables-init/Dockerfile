FROM alpine:3.16

ENV ENVOY_USER_ID=1337
ENV ENVOY_PORT=15001
ENV PORTS_TO_IGNORE=

RUN apk add -U iptables && rm -rf /var/cache/apk/*

ADD run.sh /run.sh

ENTRYPOINT ["/run.sh"]