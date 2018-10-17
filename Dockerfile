# Building stage
FROM golang:1.11-alpine AS build-env
LABEL maintainer="Kyle Bai <kyle.b@inwinstack.com>"

ENV GOPATH "/go"
ENV PROJECT_PATH "$GOPATH/src/github.com/inwinstack/ipam-operator"

RUN apk add --no-cache git make g++ && \
  go get -u github.com/golang/dep/cmd/dep

COPY . $PROJECT_PATH
RUN cd $PROJECT_PATH && \
  make dep && \
  make && mv out/operator /tmp/operator

# Running stage
FROM alpine:3.7
COPY --from=build-env /tmp/operator /bin/operator
ENTRYPOINT ["operator"]
