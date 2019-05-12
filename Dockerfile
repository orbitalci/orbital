FROM golang:1.12-alpine3.9 as base

RUN apk -v --update add \
  tini \
  protobuf-dev \
  bash \
  make \
  git \
  docker \
  python3-dev \
  py3-pip \
  gcc \
  musl-dev \
  libffi-dev \
  openssl-dev && \
    go get -v -u github.com/golang/protobuf/proto && \
    go get -v -u github.com/golang/protobuf/protoc-gen-go && \
    go get -v -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway && \
    go get -v -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger && \
    go get -v -u github.com/favadi/protoc-go-inject-tag && \
  rm /var/cache/apk/* && \
  pip3 install docker-compose

FROM base as builder

WORKDIR /orbital/
COPY . .
RUN make proto && make local-release

FROM alpine:3.9

COPY --from=builder /go/bin/admin /usr/local/bin/
COPY --from=builder /go/bin/changecheck /usr/local/bin/
COPY --from=builder /go/bin/hookhandler /usr/local/bin/
COPY --from=builder /go/bin/ocelot /usr/local/bin/
COPY --from=builder /go/bin/poller /usr/local/bin/
COPY --from=builder /go/bin/werker /usr/local/bin/
COPY --from=builder /sbin/tini /sbin/

ENTRYPOINT ["/sbin/tini", "--"]