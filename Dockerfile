FROM golang:1.20.2-alpine3.17 AS build

RUN apk add build-base

WORKDIR /usr/src/gort

COPY go.mod ./
COPY go.sum ./
COPY internal ./internal
COPY cmd ./cmd

RUN go build -ldflags="-s -w" -o /usr/local/bin/gort cmd/gort/main.go

FROM alpine:3.17

ENV GIN_MODE=release

WORKDIR /data

COPY --from=build /usr/local/bin/gort /usr/local/bin/gort

CMD ["/usr/local/bin/gort"]
