tonic::include_proto!("orbital_types");

impl From<i32> for JobTrigger {
    fn from(job_trigger: i32) -> Self {
        match job_trigger {
            0 => Self::Unspecified,
            1 => Self::Push,
            2 => Self::PullRequest,
            3 => Self::Webhook,
            4 => Self::Poll,
            5 => Self::Manual,
            6 => Self::SubscribeTrigger,
            7 => Self::CommitMsgTrigger,
            _ => panic!("Unknown job trigger"),
        }
    }
}

impl From<i32> for JobState {
    fn from(job_state: i32) -> Self {
        match job_state {
            0 => Self::Unspecified,
            1 => Self::Unknown,
            2 => Self::Queued,
            3 => Self::Starting,
            4 => Self::Running,
            5 => Self::Finishing,
            6 => Self::Cancelled,
            7 => Self::SystemErr,
            8 => Self::Failed,
            9 => Self::Done,
            10 => Self::Deleted,
            _ => panic!("Unknown job state"),
        }
    }
}

impl From<i32> for ActiveState {
    fn from(active_state: i32) -> Self {
        match active_state {
            0 => Self::Unspecified,
            1 => Self::Enabled,
            2 => Self::Disabled,
            3 => Self::Unknown,
            4 => Self::Deleted,
            _ => panic!("Unknown active state"),
        }
    }
}

impl From<i32> for SecretType {
    fn from(secret_type: i32) -> Self {
        match secret_type {
            0 => Self::Unspecified,
            1 => Self::BasicAuth,
            2 => Self::ApiKey,
            3 => Self::EnvVar,
            4 => Self::File,
            5 => Self::SshKey,
            6 => Self::DockerRegistry,
            7 => Self::NpmRepo,
            8 => Self::PypiRegistry,
            9 => Self::MavenRepo,
            10 => Self::Kubernetes,
            _ => panic!("Unknown secret type"),
        }
    }
}

impl From<i32> for CodeHostType {
    fn from(code_host_type: i32) -> Self {
        match code_host_type {
            0 => Self::CodeHostUnspecified,
            1 => Self::Generic,
            2 => Self::Bitbucket,
            3 => Self::Github,
            _ => panic!("Unknown code host type"),
        }
    }
}

impl From<i32> for NotifyType {
    fn from(notify_type: i32) -> Self {
        match notify_type {
            0 => Self::Unspecified,
            1 => Self::Slack,
            2 => Self::Webhook,
            _ => panic!("Unknown notify type"),
        }
    }
}
