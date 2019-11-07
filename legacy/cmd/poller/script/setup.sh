#!/usr/bin/env bash
echo "Seting up rsyslog,cron,postfix for reporting"
echo "$(sh -c export)" >> /root/cronenv
chmod +x /root/cronenv
service rsyslog start
service cron start
service postfix start

/poller