# Create and manage roles
path "auth/approle/*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}

# Write ACL policies
path "sys/policy/*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}

# to manage kubernetes
path "auth/kubernetes/role/*" {
  capabilities = [ "create", "read", "update", "delete", "list" ]
}