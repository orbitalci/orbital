#!/usr/bin/env bash


if [ $# -eq 1 ]; then
  version=$1
  echo "tagging"
  docker tag ocelot-poller jessishank/ocelot-poller:${version}
  docker tag ocelot-werker jessishank/ocelot-werker:${version}
  docker tag ocelot-hookhandler jessishank/ocelot-hookhandler:${version}
  docker tag ocelot-admin jessishank/ocelot-admin:${version}
  echo "pushing"
  docker push jessishank/ocelot-poller:${version}
  docker push jessishank/ocelot-werker:${version}
  docker push jessishank/ocelot-hookhandler:${version}
  docker push jessishank/ocelot-admin:${version}
else
  echo "need an argument for tag name"
  exit 1
fi



docker tag ocelot-poller jessishank/ocelot-poller