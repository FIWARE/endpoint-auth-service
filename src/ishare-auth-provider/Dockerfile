FROM registry.access.redhat.com/ubi8/go-toolset:1.16.12-7

ENV SERVER_PORT="8080"
ENV CONFIGURATION_SERVICE_URL="http://endpoint-configuration-service:8080"
ENV CERTIFICATE_FOLDER="/go/src/app/certs"
ENV GIN_MODE=release
ENV JSON_LOGGING_ENABLED=true


WORKDIR /go/src/app
COPY ./ ./


RUN mkdir ./certs


RUN go get -d -v ./...
RUN go install -v ./...

CMD ["ishare-auth-provider"]