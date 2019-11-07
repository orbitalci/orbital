#!/bin/bash

# order of arguments: vault token, vault ssh key path
# TODO: GET VAULT FROM ENVIRONMENT VARIABLE
if [ $# -gt 0 ]; then
  echo "attempting to download an ssh key file"
  args=("$@")
  vaulttoken=${args[0]}
  vaultpath=${args[1]}
#  this line that's commented out will work if you are running consul/vault/werker/hookhandler/admin all locally
#  keyFile=$(curl --header "X-Vault-Token: ${vaulttoken}" "http://docker.for.mac.localhost:8200/v1/secret/${vaultpath}" | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["data"]["sshKey"]')
  keyFile=$(curl --header "X-Vault-Token: ${vaulttoken}" "${VAULT_ADDR}/v1/secret/${vaultpath}" | python -c 'import json,sys;obj=json.load(sys.stdin);print obj["data"]["sshKey"]')
  mkdir -p ~/.ssh
  # we have to set IFS to empty so that keyfile's newlines will be preserved
  IFS=
  echo ${keyFile} >> ~/.ssh/id_rsa
  chmod 600 ~/.ssh/id_rsa
else
  echo "no arguments were passed in"
  exit 1
fi


