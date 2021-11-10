FROM golang:1.17

ENV IDP_URL="https://idp-happypets-shop.i4trust-dev.fiware.io/oauth2/token"
ENV IDP_ID="EU.EORI.NLHAPPYPETS"
ENV CLIENT_ID="EU.EORI.NLHAPPYPETS"
ENV SERVER_PORT="8080"

WORKDIR /go/src/app
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./auth.go ./auth.go

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["auth-provider"]