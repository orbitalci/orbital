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
    'api_key',
    'basic_auth'
    'docker_registry',
    'env_var',
    'file',
    'kubernetes',
    'maven_repo',
    'npm_repo',
    'pypi_registry',
    'ssh_key',
);

CREATE TYPE notify_type AS ENUM (
    'slack',
    'webhook'
);
