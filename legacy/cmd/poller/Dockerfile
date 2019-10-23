FROM ocelot-build as builder

ARG SERVICE_NAME=poller
WORKDIR /go/src/github.com/shankj3/ocelot/
COPY . .

RUN cd cmd/${SERVICE_NAME} && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/${SERVICE_NAME} .
RUN cd cmd/changecheck && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/changecheck

FROM ubuntu:latest
ENV SERVICE_NAME poller
RUN echo "postfix postfix/mailname string your.hostname.com" | debconf-set-selections
RUN echo "postfix postfix/main_mailer_type string 'No configuration'" | debconf-set-selections
RUN apt-get update && apt-get -y install -qq --force-yes cron rsyslog postfix
RUN echo "cron.* /var/log/cron.log" >> /etc/rsyslog.conf
COPY cmd/poller/script/setup.sh /sbin/start.sh
COPY cmd/poller/script/run_changecheck.sh /bin/run_changecheck.sh
COPY cmd/poller/script/cronmailrotate /etc/logrotate.d/cronmailrotate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/changecheck /bin/changecheck
COPY --from=builder /go/bin/${SERVICE_NAME} /
COPY --from=builder /sbin/tini /sbin/
ENTRYPOINT /sbin/start.sh

