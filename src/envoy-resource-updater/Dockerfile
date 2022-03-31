FROM golang:1.18-alpine

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV CONFIG_MAP_FOLDER="/configmap-folder"
ENV RUN_AS_INIT="false"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["envoy-resource-updater"]