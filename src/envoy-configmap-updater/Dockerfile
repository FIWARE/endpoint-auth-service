FROM golang:1.18-alpine

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV PROXY_CONFIG_MAP="envoy-config"
ENV PROXY_CONFIG_MAP_NAMESPACE="envoy"


WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go install envoy-configmap-updater

CMD ["envoy-configmap-updater"]