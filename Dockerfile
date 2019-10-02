FROM golang:1.13-alpine AS build

ENV GOPATH "/go"
ENV PROJECT_PATH "$GOPATH/src/github.com/xenolog/ipam"
ENV GO111MODULE "on"

COPY . $PROJECT_PATH
RUN cd $PROJECT_PATH && \
  make && mv out/controller /tmp/controller

# Running stage
FROM alpine:3.9
COPY --from=build /tmp/controller /bin/controller
ENTRYPOINT ["controller"]