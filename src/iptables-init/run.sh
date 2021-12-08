#!/bin/sh

iptables -t nat -A OUTPUT -m owner --uid-owner $ENVOY_USER_ID -j RETURN
iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports $ENVOY_PORT