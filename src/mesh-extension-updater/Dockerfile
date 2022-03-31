FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV MESH_CONFIG_FOLDER="/mesh-config"
ENV MESH_EXTENSION_FILE_NAME="service-mesh-extension.yaml"

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./main.go ./main.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["mesh-extension-updater"]