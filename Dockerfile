FROM golang:1.12-alpine AS builder

WORKDIR /src/app

COPY go.mod go.sum ./
RUN apk add --no-cache git \
    && go mod download

COPY . .
RUN go install

FROM alpine:latest
LABEL maintainer "Alex Simenduev <shamil.si@gmail.com>"

ENTRYPOINT ["odfe-alerts-handler"]

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/odfe-alerts-handler /usr/local/bin/
