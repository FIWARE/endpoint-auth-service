FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV MESH_CONFIG_FOLDER="/mesh-config"
ENV MESH_EXTENSION_FILE_NAME="service-mesh-extension.yaml"

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./main.go ./main.go

RUN go install mesh-extension-updater

CMD ["mesh-extension-updater"]