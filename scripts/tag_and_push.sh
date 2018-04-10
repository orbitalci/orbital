#!/usr/bin/env bash


if [ $# -eq 1 ]; then
  version=$1
  echo "tagging"
  docker tag ocelot-poller docker.metaverse.l11.com/ocelot-poller:${version}
  docker tag ocelot-werker docker.metaverse.l11.com/ocelot-werker:${version}
  docker tag ocelot-hookhandler docker.metaverse.l11.com/ocelot-hookhandler:${version}
  docker tag ocelot-admin docker.metaverse.l11.com/ocelot-admin:${version}
#  docker tag ocelot-build docker.metaverse.l11.com/ocelot-build:${version}
  echo "pushing"
  docker push docker.metaverse.l11.com/ocelot-poller:${version}
  docker push docker.metaverse.l11.com/ocelot-werker:${version}
  docker push docker.metaverse.l11.com/ocelot-hookhandler:${version}
  docker push docker.metaverse.l11.com/ocelot-admin:${version}
#  docker push docker.metaverse.l11.com/ocelot-build:${version}
else
  echo "need an argument for tag name"
  exit 1
fi

echo ${version} > .version