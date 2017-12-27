#!/bin/bash

# order of arguments: BBTOKEN, BBDOWNLOAD PATH, GIT COMMIT
# todo: make sure unzip is installed
# todo: handle sigterm gracefully, after this container should shut down

if [ $# -gt 0 ]; then
  count=0
  args=("$@")

  download_url=${args[1]}/${args[2]}.zip
  wget --header="Authorization: Bearer ${args[0]}" "${download_url}"
  mkdir ${args[2]}

  codedir=$(unzip ${args[2]}.zip | awk 'NR==3 {print $2}')
  cp -r ${codedir}. /${args[2]}
  # cleanup
  rm -rf ${codedir}
  rm ${args[2]}.zip

  echo "Finished with downloading source code"
  sleep infinity
else
    echo "no arguments were passed in"
fi





