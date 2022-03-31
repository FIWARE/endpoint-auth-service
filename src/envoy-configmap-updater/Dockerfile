FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV PROXY_CONFIG_MAP="envoy-config"
ENV PROXY_CONFIG_MAP_NAMESPACE="envoy"

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["envoy-configmap-updater"]