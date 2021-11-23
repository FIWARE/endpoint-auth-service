#!/bin/bash

cp -a /initial-config/. /etc/envoy/

./docker-entrypoint.sh -l trace -c /etc/envoy/envoy.yaml