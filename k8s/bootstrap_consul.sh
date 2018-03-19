consul kv put config/ocelot/postgres/db postgres
consul kv put config/ocelot/postgres/location 10.1.72.187
consul kv put config/ocelot/postgres/port 5432
consul kv put config/ocelot/postgres/username postgres
consul kv put config/ocelot/storagetype postgres

vault write secret/config/ocelot/postgres clientsecret=alUopwFppZ