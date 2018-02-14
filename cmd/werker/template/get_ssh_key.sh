#!/bin/bash

# order of arguments: vault token, vault ssh key path

if [ $# -gt 0 ]; then
  echo "first we're gonna install jq..."
  apt-get install -y jq
  echo "attempting to download an ssh key file"
  args=("$@")
  vaulttoken=${args[0]}
  vaultpath=${args[1]}
  keyFile=$(curl --header "X-Vault-Token: ${vaulttoken}" "http://127.0.0.1:8200/v1/secret/${vaultpath}" | jq -r '."data".sshKey')
  echo ${keyFile} >> ~/.ssh/id_rsa
else
  echo "no arguments were passed in"
  exit 1
fi


