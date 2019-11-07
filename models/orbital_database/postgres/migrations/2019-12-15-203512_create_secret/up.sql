CREATE TABLE secret (
  id SERIAL PRIMARY KEY,
  org_id INTEGER REFERENCES org(id),
  name TEXT NOT NULL,
  secret_type secret_type NOT NULL,
  vault_path TEXT NOT NULL,
  active_state active_state NOT NULL DEFAULT 'enabled'::active_state
);
