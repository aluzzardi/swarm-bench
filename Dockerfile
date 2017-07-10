FROM golang:1.8.3-alpine

WORKDIR /go/src/github.com/docker/aluzzardi/swarm-bench
COPY . /go/src/github.com/docker/aluzzardi/swarm-bench

RUN CGO_ENABLED=0 go install -v -ldflags="-s -w"

FROM alpine:3.6

RUN apk --no-cache add ca-certificates

COPY --from=0 /go/bin/swarm-bench /usr/local/bin/swarm-bench

ENTRYPOINT ["swarm-bench"]
