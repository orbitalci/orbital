<<<<<<< Updated upstream:Dockerfile.build
FROM golang:1.11-alpine as builder

RUN apk --update --no-cache add git protobuf-dev ca-certificates openssh python mercurial tini && \
    go get -v -u github.com/golang/dep/cmd/dep && \
=======
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
>>>>>>> Stashed changes:Dockerfile
    go get -v -u github.com/golang/protobuf/proto && \
    go get -v -u github.com/golang/protobuf/protoc-gen-go && \
    go get -v -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway && \
    go get -v -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger && \
<<<<<<< Updated upstream:Dockerfile.build
    go get -v -u github.com/favadi/protoc-go-inject-tag

WORKDIR /go/src/github.com/level11consulting/ocelot/
COPY . .
#RUN cd models && ./build-protos.sh && cd -
#RUN make proto

RUN dep ensure -v
RUN apk --update --no-cache add openssl wget bash zip curl curl-dev docker
RUN apk -v --update add \
        python \
        py-pip \
        groff \
        less \
        mailcap \
        make \
        gcc \
        libc-dev \
        && \
    pip install --upgrade awscli==1.14.5 s3cmd==2.0.1 python-magic docker-compose && \
    apk -v --purge del py-pip && \
    rm /var/cache/apk/*
=======
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
>>>>>>> Stashed changes:Dockerfile
