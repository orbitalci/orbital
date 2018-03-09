#!/usr/bin/env sh

# order of arguments: DOCKER_CONFIG_JSON (base64 encoded to avoid any weird escaping issues)
if [ $# -gt 0 ]; then
  dockersettings=$1
  if [ ! -z "${dockersettings}" ]; then
    mkdir -p ~/.docker/
    echo ${dockersettings} | base64 -d > ~/.docker/config.json
  else
    echo "docker settings var empty, saving nothing to ~/.docker/config.json"
    exit 1
  fi
else
    echo "no arguments were passed in"
    exit 1
fi