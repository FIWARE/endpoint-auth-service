FROM golang:1.18-alpine

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV CONFIG_MAP_FOLDER="/configmap-folder"
ENV RUN_AS_INIT="false"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go install envoy-resource-updater

CMD ["envoy-resource-updater"]