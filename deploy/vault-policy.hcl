# List, create, update, and delete key/value secrets
path "secret/data/config/ocelot/*"
{
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/data/creds/*"
{
  capabilities = ["create", "read", "update", "delete", "list"]
}
NewEnvAuthClient