FROM orbitalci-builder as builder

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
COPY --from=builder /orbital/scripts/docker-entrypoint.sh /usr/local/sbin/

ENTRYPOINT ["tini", "--"]
