FROM ocelot-build as builder

ARG SERVICE_NAME=hookhandler
WORKDIR /go/src/bitbucket.org/level11consulting/ocelot/
COPY . .

RUN cd cmd/${SERVICE_NAME} && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/${SERVICE_NAME} .

FROM alpine:3.7
ENV SERVICE_NAME hookhandler
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/${SERVICE_NAME} /
COPY --from=builder /sbin/tini /sbin/
CMD ["/sbin/tini", "--", "/hookhandler"]
