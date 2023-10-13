FROM golang:1.20.9-alpine

ENV MESH_CONFIG_FOLDER="/mesh-config"
ENV MESH_EXTENSION_FILE_NAME="service-mesh-extension.yaml"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./main.go ./main.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["mesh-extension-updater"]