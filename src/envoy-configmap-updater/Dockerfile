FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV PROXY_CONFIG_MAP="envoy-config"
ENV PROXY_CONFIG_MAP_NAMESPACE="envoy"

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go install envoy-configmap-updater

CMD ["envoy-configmap-updater"]