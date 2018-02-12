#!/bin/bash

# order of arguments: BBTOKEN, BBDOWNLOAD PATH, GIT COMMIT
# todo: make sure unzip is installed
# todo: handle sigterm gracefully, after this container should shut down

if [ $# -gt 0 ]; then
  count=0
  args=("$@")
  bbtoken=${args[0]}
  gitclonepath=${args[1]}
  commit=${args[2]}
  git clone ${gitclonepath} /${commit}
  echo "cloned repo to /${commit}"
  cd /${commit}
  git checkout ${commit}
  echo "Ocelot has finished with downloading source code"
  while true; do sleep 3800; done
else
    echo "no arguments were passed in"
    exit 1
fi





