#!/bin/bash

# this will setup consul + vault for postgres and assumes you've set the VAULT_TOKEN env variable
consul kv put config/ocelot/storagetype postgres

consul kv put config/ocelot/postgres/db postgres
consul kv put config/ocelot/postgres/localhost localhost
consul kv put config/ocelot/postgres/port 5432
consul kv put config/ocelot/postgres/username postgres

vault write secret/config/ocelot/postgres clientsecret=mysecretpassword