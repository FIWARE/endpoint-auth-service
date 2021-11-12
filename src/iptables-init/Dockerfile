FROM alpine:3.14.2

RUN apk add -U iptables && rm -rf /var/cache/apk/*

ADD run.sh /run.sh

ENTRYPOINT ["/run.sh"]