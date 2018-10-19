#!/usr/bin/env bash

POLLER=0
HOOKHANDLER=0
ADMIN=0

if [[ -z ${VERSION} ]]; then
    echo '$VERSION is a required environment variable for tagging the images appropriately'
    exit 1
fi

if [ $# -eq 0 ]
  then
    echo "no args supplied, tagging & pushing poller, hookhandler, and admin"
    POLLER=1
    HOOKHANDLER=1
    ADMIN=1
fi

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --poller)
    echo "will tag and push poller"
    POLLER=1
    shift
    ;;
    --hookhandler)
    HOOKHANDLER=1
    echo "will tag and push hookhandler"
    shift
    ;;
    --admin)
    ADMIN=1
    echo "will tag and push admin"
    shift
    ;;
    *)
    echo "unrecognized flag ${key}"
    shift
    ;;
esac
done

if (( POLLER == 1 )); then
    docker tag ocelot-poller docker.metaverse.l11.com/ocelot-poller:${VERSION}
    docker push docker.metaverse.l11.com/ocelot-poller:${VERSION}
fi

if (( HOOKHANDLER == 1 )); then
    docker tag ocelot-hookhandler docker.metaverse.l11.com/ocelot-hookhandler:${VERSION}
    docker push docker.metaverse.l11.com/ocelot-hookhandler:${VERSION}
fi

if (( ADMIN == 1 )); then
    docker tag ocelot-admin docker.metaverse.l11.com/ocelot-admin:${VERSION}
    docker push docker.metaverse.l11.com/ocelot-admin:${VERSION}
fi

#
#
#if [ $# -eq 1 ]; then
#  version=$1
#  echo "tagging"
#  docker tag ocelot-poller docker.metaverse.l11.com/ocelot-poller:${version}
##  docker tag ocelot-werker docker.metaverse.l11.com/ocelot-werker:${version}
#  docker tag ocelot-hookhandler docker.metaverse.l11.com/ocelot-hookhandler:${version}
#  docker tag ocelot-admin docker.metaverse.l11.com/ocelot-admin:${version}
##  docker tag ocelot-build docker.metaverse.l11.com/ocelot-build:${version}
#  echo "pushing"
#  docker push docker.metaverse.l11.com/ocelot-poller:${version}
##  docker push docker.metaverse.l11.com/ocelot-werker:${version}
#  docker push docker.metaverse.l11.com/ocelot-hookhandler:${version}
#  docker push docker.metaverse.l11.com/ocelot-admin:${version}
##  docker push docker.metaverse.l11.com/ocelot-build:${version}
#else
#  echo "need an argument for tag name"
#  exit 1
#fi

#echo ${VERSION} > .version
