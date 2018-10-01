FROM ocelot-build as builder

ARG SERVICE_NAME=hookhandler
WORKDIR /go/src/github.com/shankj3/ocelot/
COPY . .

RUN make static-linux-bin SERVICE_NAME=${SERVICE_NAME}

FROM alpine:3.8
ENV SERVICE_NAME hookhandler
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/${SERVICE_NAME} /
COPY --from=builder /sbin/tini /sbin/
CMD ["/sbin/tini", "--", "/hookhandler"]
