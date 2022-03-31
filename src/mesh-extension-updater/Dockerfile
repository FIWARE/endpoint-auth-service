FROM golang:1.18-alpine

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV MESH_CONFIG_FOLDER="/mesh-config"
ENV MESH_EXTENSION_FILE_NAME="service-mesh-extension.yaml"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./main.go ./main.go

RUN go install mesh-extension-updater

CMD ["mesh-extension-updater"]