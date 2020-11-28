FROM golang:1-alpine AS builder

LABEL maintainer=orginux

WORKDIR ${GOPATH}/src/gitlab-tool/

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/gitlab-tool

FROM alpine:3.12
COPY --from=builder /go/bin/gitlab-tool /usr/local/bin
