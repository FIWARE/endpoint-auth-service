FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV SERVER_PORT="8080"
ENV CONFIGURATION_SERVICE_URL="http://endpoint-configuration-service:8080"
ENV CERTIFICATE_FOLDER="/certs"
ENV GIN_MODE=release
ENV JSON_LOGGING_ENABLED=true

USER root

RUN mkdir /certs
RUN chmod a+rw /certs

USER 1001

WORKDIR /go/src/app
COPY ./ ./

RUN go get
RUN go install ishare-auth-provider


CMD ["ishare-auth-provider"]