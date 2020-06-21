//use orbital_headers::orbital_types::JobTrigger;
use anyhow::Result;
use chrono::NaiveDateTime;
use log::debug;
use machine::{machine, transitions};
use orbital_database::postgres;
use orbital_database::postgres::build_summary::NewBuildSummary;
use orbital_database::postgres::schema::JobTrigger;

machine! {
    #[derive(Clone, Debug, PartialEq)]
    enum BuildState {
        Queued,
        Starting,
        Running,
        Finishing,
        Done,
        Cancelled,
        Failed,
        SystemErr
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct Step;

transitions!(BuildState,
    [
        (Queued, Step) => [Starting, Cancelled, SystemErr],
        (Starting, Step) => [Running, Cancelled, SystemErr],
        (Running, Step) => [Running, Finishing, Cancelled, SystemErr],
        (Finishing, Step) => [Failed, Done]
    ]
);

impl Queued {
    pub fn on_step(self, input: Step) -> BuildState {
        BuildState::starting()
    }
}

impl Starting {
    pub fn on_step(self, input: Step) -> BuildState {
        BuildState::running()
    }
}

impl Running {
    pub fn on_step(self, input: Step) -> BuildState {
        BuildState::finishing()
    }
}

impl Finishing {
    pub fn on_step(self, input: Step) -> BuildState {
        BuildState::done()
    }
}

//#[derive(Debug)]
//pub struct BuildStateQueued;
//#[derive(Debug)]
//pub struct BuildStateStarting;
//#[derive(Debug)]
//pub struct BuildStateRunning;
//#[derive(Debug)]
//pub struct BuildStateCanceled;
//#[derive(Debug)]
//pub struct BuildStateSystemErr;
//#[derive(Debug)]
//pub struct BuildStateFailed;
//#[derive(Debug)]
//pub struct BuildStateDone;

#[derive(Debug)]
pub struct BuildContext {
    pub org: String,
    pub repo: String,
    pub branch: String,
    pub id: Option<i32>,
    pub hash: Option<String>,
    pub user_envs: Option<Vec<String>>,
    pub job_trigger: JobTrigger,
    pub queue_time: Option<NaiveDateTime>,
    _state: BuildState,
}

impl BuildContext {
    pub fn new() -> Self {
        BuildContext {
            org: "".to_string(),
            repo: "".to_string(),
            branch: "".to_string(),
            id: None,
            hash: None,
            user_envs: None,
            job_trigger: JobTrigger::Manual,
            queue_time: None,
            _state: BuildState::queued(),
        }
    }

    pub fn org(mut self, org: String) -> BuildContext {
        self.org = org;
        self
    }

    pub fn repo(mut self, repo: String) -> BuildContext {
        self.repo = repo;
        self
    }

    pub fn branch(mut self, branch: String) -> BuildContext {
        self.branch = branch;
        self
    }

    pub fn id(mut self, id: i32) -> BuildContext {
        self.id = Some(id);
        self
    }

    pub fn hash(mut self, hash: String) -> BuildContext {
        self.hash = Some(hash);
        self
    }

    pub fn triggered_by(mut self, trigger: JobTrigger) -> BuildContext {
        self.job_trigger = trigger;
        self
    }

    pub fn queue(mut self) -> Result<BuildContext> {
        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        // Add build target record in db
        debug!("Adding new build target to DB");
        let build_target_result = postgres::client::build_target_add(
            &pg_conn,
            &self.org,
            &self.repo,
            &self.hash.clone().expect("No repo hash to target"),
            &self.branch,
            Some(self.user_envs.clone().unwrap_or_default().join("")),
            self.job_trigger.clone(),
        )
        .expect("Build target add failed");

        let (_org_db, _repo_db, build_target_db) = (
            build_target_result.0,
            build_target_result.1,
            build_target_result.2,
        );

        // Add the build id and queue timestamp BuildContext
        self.id = Some(build_target_db.build_index);
        self.queue_time = Some(build_target_db.queue_time);

        // Create a new build summary record
        debug!("Adding new build summary to DB");
        let _build_summary_result_add = postgres::client::build_summary_add(
            &pg_conn,
            &self.org,
            &self.repo,
            &self.hash.clone().expect("No repo hash to target"),
            &self.branch,
            self.id.clone().expect("No build id defined"),
            NewBuildSummary {
                build_target_id: build_target_db.id,
                build_state: postgres::schema::JobState::Queued,
                start_time: None,
                ..Default::default()
            },
        )
        .expect("Unable to create new build summary");

        Ok(self)
    }

}