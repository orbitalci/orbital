#!/bin/bash

# starts infrastructure needed for ocelot, can optionally disable consul/vault with --no-consul, --no-vault, --no-postgres, --no-nexus or --no-nsq flags
args=()
follow=(" -d")
CONSUL=1
VAULT=1
NSQ=1
NSQTJ=0
POSTGRES=1
NEXUS=0
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -f|--follow)
    follow=()
    shift
    ;;
    --no-consul)
    echo "starting without consul"
    CONSUL=0
    shift
    ;;
    --no-vault)
    VAULT=0
    echo "starting without vault"
    shift
    ;;
    --no-postgres)
    POSTGRES=0
    echo "starting without postgres"
    shift
    ;;
    --no-nsq)
    NSQ=0
    echo "starting without NSQ"
    shift
    ;;
    --tj)
    NSQTJ=0
    NSQTJ=1
    echo "starting infra with docker network"
    shift
    ;;
    --no-nexus)
    NEXUS=0
    echo "starting without NEXUS"
    shift
    ;;
    *)
    echo "unrecognized flag ${key}"
    shift
    ;;
esac
done

if (( CONSUL == 1 )); then
    args+=(" -f deploy/infra/consul-docker-compose.yml")
fi

if (( VAULT == 1 )); then
    args+=(" -f deploy/infra/vault-docker-compose.yml")
fi

if (( POSTGRES == 1 )); then
    args+=(" -f deploy/infra/postgres-docker-compose.yml")
fi

if (( NSQ == 1 )); then
    args+=(" -f deploy/infra/nsq-docker-compose.yml")
fi

if (( NSQTJ == 1 )); then
    args+=(" -f deploy/infra/nsq-docker-compose-tj.yml")
fi

if (( NEXUS == 1 )); then
    args+=(" -f deploy/infra/nexus-docker-compose.yml")
fi

docker-compose${args[@]} up${follow[@]}
# TODO: TJ has some stuff that will unseal vault for us! Perhaps it goes here?
