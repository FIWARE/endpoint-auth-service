#!/bin/sh

iptables -t nat -A OUTPUT -m owner --uid-owner 1337 -j RETURN
iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports 15001