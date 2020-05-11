CREATE TABLE repo (
  id SERIAL PRIMARY KEY,
  org_id INTEGER REFERENCES org(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  uri TEXT NOT NULL,
  git_host_type git_host_type NOT NULL,
  secret_id INTEGER REFERENCES secret(id),
  build_active_state active_state NOT NULL DEFAULT 'enabled'::active_state,
  notify_active_state active_state NOT NULL DEFAULT 'enabled'::active_state,
  next_build_index INTEGER NOT NULL DEFAULT 1
);
