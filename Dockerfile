# Building stage
FROM inwinstack/golang:1.11-alpine AS build-env
LABEL maintainer="Kyle Bai <kyle.b@inwinstack.com>"

ENV GOPATH "/go"
ENV PROJECT_PATH "$GOPATH/src/github.com/inwinstack/ipam"

COPY . $PROJECT_PATH
RUN cd $PROJECT_PATH && \
  make dep && \
  make && mv out/controller /tmp/controller

# Running stage
FROM alpine:3.7
COPY --from=build-env /tmp/controller /bin/controller
ENTRYPOINT ["controller"]
