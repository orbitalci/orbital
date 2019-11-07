CREATE TABLE build_summary (
  id SERIAL PRIMARY KEY,
  build_target_id INTEGER REFERENCES build_target(id),
  start_time TIMESTAMP,
  end_time TIMESTAMP,
  build_state job_state NOT NULL
);