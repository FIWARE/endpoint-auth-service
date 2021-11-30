#!/bin/bash

cp -a /initial-config/. /etc/envoy/

./docker-entrypoint.sh -l debug -c /etc/envoy/envoy.yaml