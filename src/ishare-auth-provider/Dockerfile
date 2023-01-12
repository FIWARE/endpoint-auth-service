FROM golang:1.19.2-alpine

ENV SERVER_PORT="8080"
ENV CONFIGURATION_SERVICE_URL="http://endpoint-configuration-service:8080"
ENV CERTIFICATE_FOLDER="/certs"
ENV GIN_MODE=release
ENV JSON_LOGGING_ENABLED=true

RUN mkdir /certs

WORKDIR /go/src/app
COPY ./ ./

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["ishare-auth-provider"]

