#!/bin/sh

for port in $(echo $PORTS_TO_IGNORE | tr "," "\n") 
do
    iptables -t nat -A OUTPUT -p tcp --dport $port -j RETURN
done

iptables -t nat -A OUTPUT -m owner --uid-owner $ENVOY_USER_ID -j RETURN
iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports $ENVOY_PORT