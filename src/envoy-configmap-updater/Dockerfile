FROM golang:1.21.11-alpine

ENV PROXY_CONFIG_FOLDER="/proxy-config"
ENV PROXY_CONFIG_MAP="envoy-config"
ENV PROXY_CONFIG_MAP_NAMESPACE="envoy"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["envoy-configmap-updater"]