tonic::include_proto!("orbital_types");

use std::fmt;
use std::str::FromStr;

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

impl From<String> for JobTrigger {
    fn from(job_trigger: String) -> Self {
        match job_trigger.to_lowercase().as_str() {
            "unspecified" => Self::Unspecified,
            "push" => Self::Push,
            "pullrequest" => Self::PullRequest,
            "webhook" => Self::Webhook,
            "poll" => Self::Poll,
            "manual" => Self::Manual,
            "subscribetrigger" => Self::SubscribeTrigger,
            "commitmsgtrigger" => Self::CommitMsgTrigger,
            _ => panic!("Unknown job trigger"),
        }
    }
}

impl fmt::Display for JobTrigger {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Unspecified => write!(f, "{}", "Unspecified"),
            Self::Push => write!(f, "Push"),
            Self::PullRequest => write!(f, "{}", "PullRequest"),
            Self::Webhook => write!(f, "{}", "Webhook"),
            Self::Poll => write!(f, "{}", "Poll"),
            Self::Manual => write!(f, "{}", "Manual"),
            Self::SubscribeTrigger => write!(f, "{}", "SubscribeTrigger"),
            Self::CommitMsgTrigger => write!(f, "{}", "CommitMessage"),
        }
    }
}

//impl FromStr for JobTrigger {
//    type Err = std::string::ParseError;
//
//    fn from_str(s: &str) -> Result<Self, Self::Err> {
//        Ok(match s.to_lowercase().as_ref() {
//            "unspecified" => ActiveState::Unspecified,
//            "unknown" => ActiveState::Unknown,
//            "enabled" => ActiveState::Enabled,
//            "disabled" => ActiveState::Disabled,
//            "deleted" => ActiveState::Deleted,
//            _ => ActiveState::Unknown,
//        })
//    }
//}

impl JobTrigger {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 8] {
        [
            "unspecified",
            "push",
            "pullrequest",
            "webhook",
            "poll",
            "manual",
            "subscribetrigger",
            "commitmsgtrigger",
        ]
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
            6 => Self::Canceled,
            7 => Self::SystemErr,
            8 => Self::Failed,
            9 => Self::Done,
            10 => Self::Deleted,
            _ => panic!("Unknown job state"),
        }
    }
}

impl From<String> for JobState {
    fn from(job_state: String) -> Self {
        match job_state.to_lowercase().as_str() {
            "unspecified" => Self::Unspecified,
            "unknown" => Self::Unknown,
            "queued" => Self::Queued,
            "starting" => Self::Starting,
            "running" => Self::Running,
            "finishing" => Self::Finishing,
            "canceled" => Self::Canceled,
            "systemerr" => Self::SystemErr,
            "failed" => Self::Failed,
            "done" => Self::Done,
            "deleted" => Self::Deleted,
            _ => panic!("Unknown job state"),
        }
    }
}

impl fmt::Display for JobState {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Unspecified => write!(f, "{}", "Unspecified"),
            Self::Unknown => write!(f, "{}", "Unknown"),
            Self::Queued => write!(f, "{}", "Queued"),
            Self::Starting => write!(f, "{}", "Starting"),
            Self::Running => write!(f, "{}", "Running"),
            Self::Finishing => write!(f, "{}", "Finishing"),
            Self::Canceled => write!(f, "{}", "Canceled"),
            Self::SystemErr => write!(f, "{}", "SystemErr"),
            Self::Failed => write!(f, "{}", "Failed"),
            Self::Done => write!(f, "{}", "Done"),
            Self::Deleted => write!(f, "{}", "Deleted"),
        }
    }
}

impl JobState {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 11] {
        [
            "unspecified",
            "unknown",
            "queued",
            "starting",
            "running",
            "finishing",
            "canceled",
            "systemerr",
            "failed",
            "done",
            "deleted",
        ]
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

impl From<String> for ActiveState {
    fn from(active_state: String) -> Self {
        match active_state.to_lowercase().as_str() {
            "unspecified" => Self::Unspecified,
            "enabled" => Self::Enabled,
            "disabled" => Self::Disabled,
            "unknown" => Self::Unknown,
            "deleted" => Self::Deleted,
            _ => panic!("Unknown active state"),
        }
    }
}

impl fmt::Display for ActiveState {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Unspecified => write!(f, "{}", "Unspecified"),
            Self::Enabled => write!(f, "{}", "Enabled"),
            Self::Disabled => write!(f, "{}", "Disabled"),
            Self::Unknown => write!(f, "{}", "Unknown"),
            Self::Deleted => write!(f, "{}", "Deleted"),
        }
    }
}

// TODO: Get rid of lots of duplication
//impl ActiveState {
//    /// A list of possible variants in `&'static str` form
//    pub fn variants() -> [&'static str; 5] {
//        ["unspecified", "enabled", "disabled", "unknown", "deleted"]
//    }
//}

impl From<i32> for SecretType {
    fn from(secret_type: i32) -> Self {
        match secret_type {
            0 => Self::Unspecified,
            1 => Self::ApiKey,
            2 => Self::BasicAuth,
            3 => Self::DockerRegistry,
            4 => Self::EnvVar,
            5 => Self::File,
            6 => Self::Kubernetes,
            7 => Self::MavenRepo,
            8 => Self::NpmRepo,
            9 => Self::PypiRegistry,
            10 => Self::SshKey,
            _ => panic!("Unknown secret type"),
        }
    }
}

impl From<String> for SecretType {
    fn from(secret_type: String) -> Self {
        match secret_type.to_lowercase().as_str() {
            "unspecified" => Self::Unspecified,
            "api_key" => Self::ApiKey,
            "basic_auth" => Self::BasicAuth,
            "docker_registry" => Self::DockerRegistry,
            "env_var" => Self::EnvVar,
            "file" => Self::File,
            "kubernetes" => Self::Kubernetes,
            "maven_repo" => Self::MavenRepo,
            "npm_repo" => Self::NpmRepo,
            "pypi_registry" => Self::PypiRegistry,
            "ssh_key" => Self::SshKey,
            _ => panic!("Unknown secret type"),
        }
    }
}

impl fmt::Display for SecretType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Unspecified => write!(f, "{}", "Unspecified"),
            Self::BasicAuth => write!(f, "{}", "BasicAuth"),
            Self::ApiKey => write!(f, "{}", "ApiKey"),
            Self::EnvVar => write!(f, "{}", "EnvVar"),
            Self::File => write!(f, "{}", "File"),
            Self::SshKey => write!(f, "{}", "SshKey"),
            Self::DockerRegistry => write!(f, "{}", "DockerRegistry"),
            Self::NpmRepo => write!(f, "{}", "NpmRepo"),
            Self::PypiRegistry => write!(f, "{}", "PypiRegistry"),
            Self::MavenRepo => write!(f, "{}", "MavenRepo"),
            Self::Kubernetes => write!(f, "{}", "Kubernetes"),
        }
    }
}

impl FromStr for SecretType {
    type Err = std::string::ParseError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Ok(match s.to_lowercase().as_ref() {
            "unspecified" => Self::Unspecified,
            "api_key" => Self::ApiKey,
            "basic_auth" => Self::BasicAuth,
            "docker_registry" => Self::DockerRegistry,
            "env_var" => Self::EnvVar,
            "file" => Self::File,
            "kubernetes" => Self::Kubernetes,
            "maven_repo" => Self::MavenRepo,
            "npm_repo" => Self::NpmRepo,
            "pypi_registry" => Self::PypiRegistry,
            "ssh_key" => Self::SshKey,
            _ => panic!("Unknown secret type"),
        })
    }
}

impl SecretType {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 11] {
        [
            "unspecified",
            "api_key",
            "basic_auth",
            "docker_registry",
            "env_var",
            "file",
            "kubernetes",
            "maven_repo",
            "npm_repo",
            "pypi_registry",
            "ssh_key",
        ]
    }
}

impl From<i32> for GitHostType {
    fn from(git_host_type: i32) -> Self {
        match git_host_type {
            0 => Self::Unspecified,
            1 => Self::Generic,
            2 => Self::Bitbucket,
            3 => Self::Github,
            _ => panic!("Unknown git host type"),
        }
    }
}

impl From<String> for GitHostType {
    fn from(git_host_type: String) -> Self {
        match git_host_type.to_lowercase().as_str() {
            "unspecified" => Self::Unspecified,
            "generic" => Self::Generic,
            "bitbucket" => Self::Bitbucket,
            "github" => Self::Github,
            _ => panic!("Unknown git host type"),
        }
    }
}

impl fmt::Display for GitHostType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Unspecified => write!(f, "{}", "Unspecified"),
            Self::Generic => write!(f, "{}", "Generic"),
            Self::Bitbucket => write!(f, "{}", "Bitbucket"),
            Self::Github => write!(f, "{}", "Github"),
        }
    }
}

impl FromStr for GitHostType {
    type Err = std::string::ParseError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Ok(match s.to_lowercase().as_ref() {
            "unspecified" => Self::Unspecified,
            "generic" => Self::Generic,
            "bitbucket" => Self::Bitbucket,
            "github" => Self::Github,
            _ => panic!("Unknown git host type"),
        })
    }
}

impl GitHostType {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 4] {
        ["unspecified", "generic", "bitbucket", "github"]
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

impl From<String> for NotifyType {
    fn from(notify_type: String) -> Self {
        match notify_type.to_lowercase().as_str() {
            "unspecified" => Self::Unspecified,
            "slack" => Self::Slack,
            "webhook" => Self::Webhook,
            _ => panic!("Unknown notify type"),
        }
    }
}
impl fmt::Display for NotifyType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Unspecified => write!(f, "{}", "Unspecified"),
            Self::Slack => write!(f, "{}", "Slack"),
            Self::Webhook => write!(f, "{}", "Webhook"),
        }
    }
}

impl FromStr for NotifyType {
    type Err = std::string::ParseError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        Ok(match s.to_lowercase().as_ref() {
            "unspecified" => Self::Unspecified,
            "slack" => Self::Slack,
            "webhook" => Self::Webhook,
            _ => panic!("Unknown notify type"),
        })
    }
}

impl NotifyType {
    /// A list of possible variants in `&'static str` form
    pub fn variants() -> [&'static str; 3] {
        ["unspecified", "slack", "webhook"]
    }
}
