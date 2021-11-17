FROM golang:1.17.3-alpine

ENV PROXY_CONFIG_FOLDER="/proxy-config"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./updater.go ./updater.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["envoy-configmap-updater"]