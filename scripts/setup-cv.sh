#!/bin/bash
set -e
# this will setup consul + vault for postgres and assumes you've set the VAULT_TOKEN env variable
if [ "${PATH_PREFIX}" != "" ]; then
    prefix="${PATH_PREFIX}/"
fi
echo "prefix is \"${prefix}\""

# If DBHOST is unset, check if it is passed as arg. Fallback to localhost
if [ -z "${DBHOST}" ]; then
  DBHOST=$1
  if [ "${DBHOST}" == "" ]; then
      echo "using localhost as db host location"
      DBHOST=localhost
  fi
fi

#exitcode=vault version | grep v10
#if [ $(vault version  | grep ) ]

locals() {
    echo "setting  up local"
    consul kv put ${prefix}config/ocelot/storagetype postgres
    consul kv put ${prefix}config/ocelot/postgres/db postgres
    consul kv put ${prefix}config/ocelot/postgres/location ${DBHOST}
    consul kv put ${prefix}config/ocelot/postgres/port 5432
    consul kv put ${prefix}config/ocelot/postgres/username postgres

    # Vault kv secret engine (Static secret)
    vault kv put secret/data/config/ocelot/postgres clientsecret="mysecretpassword"

    # Vault database secret engine (Dynamic secret)
    vault secrets enable database
    vault write database/config/my-postgresql-database \
        plugin_name=postgresql-database-plugin \
        allowed_roles="ocelot" \
        connection_url="postgresql://{{username}}:{{password}}@192.168.56.78:5432/?sslmode=disable" \
        username="postgres" \
        password="mysecretpassword"

    # Short TTLs, so we can experience token expiration/renewal more often
    vault write database/roles/ocelot \
        db_name=my-postgresql-database \
        creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
            GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
        default_ttl="10m" \
        max_ttl="1h"

    vault read database/creds/ocelot
}

if [ "${DEV_K8S}" == "" ]; then
    echo "setting up local"
    locals
fi


dev_k8s() {
    echo "setting up dev k8s"
    loc=$(kubectl get svc -n pedev pgsql-postgresql -o yaml -o jsonpath='{.spec.externalIPs[0]}')
    port=$(kubectl get svc -n pedev pgsql-postgresql -o yaml -o jsonpath='{.spec.ports[0].port}')
    pw=$(kubectl get secret --namespace pedev pgsql-postgresql -o jsonpath="{.data.postgres-password}" | base64 --decode; echo)
    consul kv put ${prefix}config/ocelot/storagetype postgres
    consul kv put ${prefix}config/ocelot/postgres/db ocelot
    consul kv put ${prefix}config/ocelot/postgres/location ${loc}
    consul kv put ${prefix}config/ocelot/postgres/port ${port}
    consul kv put ${prefix}config/ocelot/postgres/username ocelot
    vault kv put secret/${prefix}config/ocelot/postgres clientsecret=${pw}
}

if [ "${DEV_K8S}" != "" ]; then
    echo "setting up dev k8s in namespace pedev"
    dev_k8s
fi