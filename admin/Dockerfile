FROM golang:1.9-alpine

ENV SERVICE_NAME admin 
WORKDIR /go/src/github.com/shankj3/ocelot/
COPY . .

RUN apk --update --no-cache add git protobuf-dev && \
    go get -u github.com/golang/dep/cmd/dep && \
    go get -v -u github.com/golang/protobuf/protoc-gen-go && \
    rm -f ./protos/out/* && cd protos && ./build-protos.sh && cd .. && \
    dep ensure -v && \
    cd ${SERVICE_NAME} && go-wrapper install

WORKDIR /go/src/github.com/shankj3/ocelot/${SERVICE_NAME}
CMD ["go-wrapper", "run"]
