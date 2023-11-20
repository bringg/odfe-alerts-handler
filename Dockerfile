FROM golang:1.21-alpine AS builder

WORKDIR /src/app

COPY go.mod go.sum ./
RUN apk add --no-cache git \
    && go mod download

COPY . .
RUN go install

FROM alpine:3.18
LABEL maintainer "Alex Simenduev <shamil.si@gmail.com>"

ENTRYPOINT ["odfe-alerts-handler"]

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/bin/odfe-alerts-handler /usr/local/bin/
