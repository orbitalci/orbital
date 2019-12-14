CREATE TYPE active_state AS ENUM (
    'unknown',
    'enabled',
    'disabled'
);

CREATE TYPE job_state AS ENUM (
    'unknown',
    'queued',
    'starting',
    'running',
    'finishing',
    'canceled',
    'systemerr',
    'failed',
    'done'
);

CREATE TYPE job_trigger AS ENUM (
    'push',
    'pullrequest',
    'webhook',
    'poll',
    'manual',
    'subscribe_trigger',
    'commit_msg_trigger'
);

CREATE TYPE git_host_type AS ENUM (
    'generic',
    'bitbucket',
    'github'
);

CREATE TYPE secret_type AS ENUM (
    'docker_registry',
    'npm_repo',
    'maven_repo',
    'ssh_key',
    'helm_repo',
    'kubernetes',
    'apple_developer',
    'env_var',
    'file',
    'basic_auth'
);

CREATE TYPE notify_type AS ENUM (
    'slack',
    'webhook'
);
