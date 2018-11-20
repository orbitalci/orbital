#!/usr/bin/env bash

echo "VAULT_ADDR is $VAULT_ADDR"

if [ -z $DEPLOY_ENV ]; then
  echo "no DEPLOY_ENV set, required for binding service account to vault policy"
  exit 1
fi

vault write auth/kubernetes/role/ocelot \
  bound_service_account_names=ocelot \
  bound_service_account_namespaces=${DEPLOY_ENV} \
  policies=ocelot \
  ttl=24h
