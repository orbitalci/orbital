CREATE TYPE active_state AS ENUM (
    'UNKNOWN',
    'ENABLED',
    'DISABLED'
);

CREATE TYPE job_state AS ENUM (
    'UNKNOWN',
    'QUEUED',
    'STARTING',
    'RUNNING',
    'FINISHING',
    'CANCELED',
    'KILLED',
    'FAILED',
    'DONE'
);

CREATE TYPE job_trigger AS ENUM (
    'POLL',
    'WEBHOOK',
    'MANUAL',
    'SUBSCRIBE',
    'COMMIT_MSG'
);

CREATE TYPE code_host_type AS ENUM (
    'BITBUCKET',
    'GITHUB'
);

CREATE TYPE secret_type AS ENUM (
    'DOCKER_REGISTRY',
    'NPM_REPO',
    'MAVEN_REPO',
    'SSH_KEY',
    'HELM_REPO',
    'KUBERNETES',
    'APPLE_DEVELOPER',
    'ENV_VAR',
    'FILE',
    'BASIC_AUTH'
);

CREATE TYPE notify_type AS ENUM (
    'SLACK',
    'WEBHOOK'
);
