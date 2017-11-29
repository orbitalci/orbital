FROM golang:1.9-alpine

ENV SERVICE_NAME hookhandler
WORKDIR /go/src/${SERVICE_NAME}
COPY . .

RUN apk --update --no-cache add git && \
    cd ${SERVICE_NAME} && \
    go get -v ./... && \
    go install -v ./... && \
    rm -rf /go/src/${SERVICE_NAME}/*

CMD ["go-wrapper", "run"]
