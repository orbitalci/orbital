#!/bin/bash
set -e
# this will setup consul + vault for postgres and assumes you've set the VAULT_TOKEN env variable
if [ "${PATH_PREFIX}" != "" ]; then
    prefix="${PATH_PREFIX}/"
fi
echo "prefix is ${prefix}"


locals() {
    echo "setting  up local"
    consul kv put ${prefix}config/ocelot/storagetype postgres
    consul kv put ${prefix}config/ocelot/postgres/db postgres
    consul kv put ${prefix}config/ocelot/postgres/location localhost
    consul kv put ${prefix}config/ocelot/postgres/port 5432
    consul kv put ${prefix}config/ocelot/postgres/username postgres
    vault write secret/${prefix}config/ocelot/postgres clientsecret=mysecretpassword
    #vault kv put secret/${prefix}config/ocelot/postgres clientsecret=mysecretpassword

}

if [ "${DEV_K8S}" == "" ]; then
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
    vault write secret/${prefix}config/ocelot/postgres clientsecret=${pw}
    #vault kv put secret/${prefix}config/ocelot/postgres clientsecret=${pw}
}

if [ "${DEV_K8S}" != "" ]; then
    dev_k8s
fi