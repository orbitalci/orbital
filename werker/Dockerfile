FROM ocelot-build as builder

ARG SERVICE_NAME=werker
WORKDIR /go/src/github.com/shankj3/ocelot/
COPY . .

RUN cd cmd/${SERVICE_NAME} && CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o /go/bin/${SERVICE_NAME} .

FROM scratch
ENV SERVICE_NAME werker
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/${SERVICE_NAME} /
CMD ["/werker"]
