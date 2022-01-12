FROM golang:1.17.3-alpine

ENV MESH_CONFIG_FOLDER="/mesh-config"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./main.go ./main.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["mesh-extension-updater"]