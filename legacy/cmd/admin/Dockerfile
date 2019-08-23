FROM orbital-builder as builder

ARG SERVICE_NAME=admin
WORKDIR /go/src/github.com/shankj3/ocelot/
COPY . .

RUN make local-service SERVICE_NAME=${SERVICE_NAME}

FROM alpine:3.8
RUN apk --update --no-cache add docker
ENV SERVICE_NAME admin
COPY models/pb/*.swagger.json /swagger/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/${SERVICE_NAME} /
COPY --from=builder /sbin/tini /sbin/
CMD ["/sbin/tini", "--", "/admin"]
