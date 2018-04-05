#!/usr/bin/env sh

# order of arguments: AWS_CREDENTIAL_CONFIG (base64 encoded to avoid any weird escaping issues)
if [ $# -gt 0 ]; then
  awsfile=$1
  if [ ! -z "${awsfile}" ]; then
    mkdir -p ~/.aws
    echo ${awsfile} | base64 -d > ~/.aws/credentials
  else
    echo "aws settings var empty, saving nothing to ~/.aws/credentials"
    exit 1
  fi
else
    echo "no arguments were passed in"
    exit 1
fi