CREATE TABLE org (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created TIMESTAMP NOT NULL,
  last_update TIMESTAMP NOT NULL,
  active_state active_state NOT NULL DEFAULT 'enabled'::active_state
);
