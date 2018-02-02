##!/bin/bash
#
## order of arguments: BBTOKEN, BBDOWNLOAD PATH, GIT COMMIT
## todo: make sure unzip is installed
## todo: handle sigterm gracefully, after this container should shut down
#
#if [ $# -gt 0 ]; then
#  old=$(pwd)
#  cd /
#  count=0
#  args=("$@")
#
#  download_url=${args[1]}/${args[2]}.zip
#  echo "wget --header=Authorization: Bearer <redacted> ${args[1]}/${args[2]}.zip"
#  wget --header="Authorization: Bearer ${args[0]}" "${download_url}"
#  mkdir ${args[2]}
#
#  echo "unzipping file ${args[2]}.zip"
#  codedir=$(unzip ${args[2]}.zip | grep creating: | head -1 | awk '{print $2}')
#  echo "copying ${codedir} to /${args[2]}"
#  cp -r ${codedir} /${args[2]}
#  # cleanup
#  echo "removing ${codedir} and ${args[2]}.zip"
#  rm -rf ${codedir}
#  rm ${args[2]}.zip
#
#  echo "Finished with downloading source code"
#  echo "returning to last pwd: ${old}"
#  cd "${old}"
#  while sleep 3600; do :; done
#else
#    echo "no arguments were passed in"
#fi
#
#
#
#
#