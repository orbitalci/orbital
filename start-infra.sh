#!/bin/bash

# starts infrastructure needed for ocelot, can optionally disable consul/vault with --no-consul, --no-vault, --no-postgres, --no-nexus or --no-nsq flags
args=()
follow=(" -d")
CONSUL=1
VAULT=1
NSQ=1
POSTGRES=1
NEXUS=1
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
    args+=(" -f infra/consul-docker-compose.yml")
fi

if (( VAULT == 1 )); then
    args+=(" -f infra/vault-docker-compose.yml")
fi

if (( POSTGRES == 1 )); then
    args+=(" -f infra/postgres-docker-compose.yml")
fi

if (( NSQ == 1 )); then
    args+=(" -f infra/nsq-docker-compose.yml")
fi

if (( NEXUS == 1 )); then
    args+=(" -f infra/nexus-docker-compose.yml")
fi

docker-compose${args[@]} up${follow[@]}
# TODO: TJ has some stuff that will unseal vault for us! Perhaps it goes here?
