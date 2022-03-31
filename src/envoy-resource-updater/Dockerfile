FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV CONFIG_MAP_FOLDER="/configmap-folder"
ENV RUN_AS_INIT="false"

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go get
RUN go install envoy-resource-updater

CMD ["envoy-resource-updater"]