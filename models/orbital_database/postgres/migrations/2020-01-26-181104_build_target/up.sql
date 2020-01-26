CREATE TABLE build_target (
  id SERIAL PRIMARY KEY,
  repo_id INTEGER REFERENCES repo(id),
  git_hash VARCHAR(40) NOT NULL,
  branch TEXT NOT NULL,
  queue_time TIMESTAMP NOT NULL,
  build_index INTEGER NOT NULL,
  trigger job_trigger NOT NULL
);

