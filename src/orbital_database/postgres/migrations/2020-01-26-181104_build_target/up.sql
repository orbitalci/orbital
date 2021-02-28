CREATE TABLE build_target (
  id SERIAL PRIMARY KEY,
  repo_id INTEGER REFERENCES repo(id) ON DELETE CASCADE,
  git_hash VARCHAR(40) NOT NULL,
  branch TEXT NOT NULL,
  user_envs TEXT,
  queue_time TIMESTAMP NOT NULL,
  build_index INTEGER NOT NULL,
  trigger job_trigger NOT NULL
);

