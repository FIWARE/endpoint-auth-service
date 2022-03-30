FROM alpine:3.14

ENV ENVOY_USER_ID=1337
ENV ENVOY_PORT=15001

RUN apk add -U iptables && rm -rf /var/cache/apk/*

ADD run.sh /run.sh

ENTRYPOINT ["/run.sh"]