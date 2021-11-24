FROM golang:1.17.3-alpine

ENV SERVER_PORT="8080"
ENV CONFIGURATION_SERVICE_URL="http://endpoint-configuration-service:8080"
ENV CERTIFICATE_FOLDER="/mounted-certs"

WORKDIR /go/src/app
COPY ./ ./

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["ishare-auth-provider"]