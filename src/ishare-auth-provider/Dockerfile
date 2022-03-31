FROM golang:1.18-alpine

ENV GOPATH=$APP_ROOT
ENV GOBIN=$APP_ROOT/bin

ENV SERVER_PORT="8080"
ENV CONFIGURATION_SERVICE_URL="http://endpoint-configuration-service:8080"
ENV CERTIFICATE_FOLDER="/certs"
ENV GIN_MODE=release
ENV JSON_LOGGING_ENABLED=true

USER root

RUN mkdir /certs
RUN chmod a+rw /certs

USER 1001

COPY ./ ./ 

RUN go install ishare-auth-provider

CMD ["ishare-auth-provider"]