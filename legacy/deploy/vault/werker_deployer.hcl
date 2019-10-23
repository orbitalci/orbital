# need to be able to generate a secret-id
path "auth/approle/role/ocelot/secret-id" {
  capabilities = [ "create", "update" ]
}