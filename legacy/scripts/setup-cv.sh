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

    # Vault kv secret engine (Static secret)
    consul kv put ${prefix}config/ocelot/postgres/vault/secretsengine kv
    consul kv put ${prefix}config/ocelot/postgres/username postgres
    vault kv put secret/data/config/ocelot/postgres clientsecret="mysecretpassword"


    # Vault database secret engine (Dynamic secret)
    # Uncomment if you want to operate using the dynamic secrets
    # TODO: We should practice development using a user w/o superuser access

    #vault secrets enable database || true
    #vault write database/config/ocelot \
    #    plugin_name=postgresql-database-plugin \
    #    allowed_roles="ocelot" \
    #    connection_url="postgresql://{{username}}:{{password}}@${DBHOST}:5432/?sslmode=disable" \
    #    username="postgres" \
    #    password="mysecretpassword"

    ## Short TTLs, so we can experience token expiration/renewal more often
    ## Assuming we are using the default docker container's superuser + public schema
    #vault write database/roles/ocelot \
    #    db_name=ocelot \
    #    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    #        GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" \
    #    default_ttl="10m" \
    #    max_ttl="1h"

    ## Example of tuning role to a more minimally scoped user using
    ##vault write database/roles/ocelot \
    ##    db_name=ocelot \
    ##    creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; \
    ##        REVOKE ALL ON SCHEMA public FROM \"{{name}}\"; \
    ##        GRANT ocelot TO \"{{name}}\";" \
    ##    default_ttl="10m" \
    ##    max_ttl="1h"

    #consul kv put ${prefix}config/ocelot/postgres/vault/secretsengine database
    #consul kv put ${prefix}config/ocelot/postgres/vault/rolename ocelot

    ## Test that we can get dynamic creds from Vault
    ## In production, you will need to define policy for read to "database/creds/ocelot"
    #vault read database/creds/ocelot

    ## Vault database secret engine END

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
    consul kv put ${prefix}config/ocelot/vault/secretbackend kv
    vault kv put secret/${prefix}config/ocelot/postgres clientsecret=${pw}
}

if [ "${DEV_K8S}" != "" ]; then
    echo "setting up dev k8s in namespace pedev"
    dev_k8s
fi
