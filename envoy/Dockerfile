FROM golang:1.18-alpine AS cert-provider
FROM tinygo/tinygo:0.26.0 as build
# this is required, since the tinygo image is very minimal and does not contain the required ssl-certs to access github
WORKDIR /
COPY --from=cert-provider /etc/ssl/certs ./etc/ssl/certs
COPY /src/cached-auth-filter /cached-auth-filter
WORKDIR /cached-auth-filter
USER root
RUN tinygo build -o cached-auth-filter.wasm -target=wasi ./main.go

FROM envoyproxy/envoy:v1.23.0

COPY --from=build /cached-auth-filter/cached-auth-filter.wasm /cached-auth-filter/cached-auth-filter.wasm
RUN chmod a+rx /cached-auth-filter/cached-auth-filter.wasm
