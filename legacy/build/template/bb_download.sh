#!/bin/bash

# order of arguments: BBTOKEN, BBDOWNLOAD PATH, GIT COMMIT, CLONEPREFIX
# todo: make sure unzip is installed
# todo: handle sigterm gracefully, after this container should shut down

if [ $# -gt 0 ]; then
  args=("$@")
  bbtoken=${args[0]}
  gitclonepath=${args[1]}
  commit=${args[2]}
  clonedir=${args[3]}
  git clone ${gitclonepath} /${clonedir}
  echo "cloned repo to /${clonedir}"
  cd /${clonedir}
  git checkout ${commit}
  echo "Finished with downloading source code"
else
    echo "no arguments were passed in"
    exit 1
fi
