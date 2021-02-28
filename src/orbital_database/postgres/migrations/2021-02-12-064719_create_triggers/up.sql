CREATE TABLE triggers (
  id SERIAL PRIMARY KEY,
  repo_id INTEGER REFERENCES repo(id) ON DELETE CASCADE,
  last_target_trigger INTEGER REFERENCES build_target(id) ON DELETE CASCADE,
  branch TEXT NOT NULL,
  repo_id_list JSONB NOT NULL,
  trigger_active_state active_state NOT NULL DEFAULT 'enabled'::active_state
);