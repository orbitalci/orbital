CREATE TABLE build_stage (
  id SERIAL PRIMARY KEY,
  build_summary_id INTEGER REFERENCES build_summary(id) ON DELETE CASCADE,
  build_host TEXT NOT NULL,
  stage_name TEXT NOT NULL,
  output TEXT,
  start_time TIMESTAMP NOT NULL,
  end_time TIMESTAMP,
  exit_code INTEGER
);