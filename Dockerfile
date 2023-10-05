FROM golang:1.20-alpine as buildbase

RUN apk add build-base git

WORKDIR /go/src/github.com/rarimo/horizon-svc

COPY . .

ENV GO111MODULE="on"
ENV CGO_ENABLED=1
ENV GOOS="linux"

RUN go mod tidy
RUN go mod vendor
RUN go build -o /usr/local/bin/horizon-svc github.com/rarimo/horizon-svc

###

FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/horizon-svc /usr/local/bin/horizon-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["horizon-svc"]
