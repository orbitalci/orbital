//use orbital_headers::orbital_types::JobTrigger;
use anyhow::Result;
use chrono::NaiveDateTime;
use config_parser::OrbitalConfig;
use git_meta::git_info;
use git_url_parse::GitUrl;
use log::{debug, error, info};
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
        Fail,
        SystemErr
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct Step;

transitions!(BuildState,
    [
        (Queued, Step) => Starting,
        (Queued, Cancelled) => Cancelled,
        (Queued, SystemErr) => SystemErr,
        (Starting, Step) => Running,
        (Starting, Cancelled) => Cancelled,
        (Starting, SystemErr) => SystemErr,
        (Running, Step) => Running,
        (Running, Finishing) => Finishing,
        (Running, Cancelled) => Cancelled,
        (Running, SystemErr) => SystemErr,
        (Finishing, Fail) => Fail,
        (Finishing, Done) => Done
    ]
);

impl Queued {
    pub fn on_step(self, input: Step) -> Starting {
        println!("Queued -> Starting");
        Starting {}
    }

    pub fn on_cancelled(self, _: Cancelled) -> Cancelled {
        println!("Queued -> Cancelled");
        Cancelled {}
    }

    pub fn on_system_err(self, _: SystemErr) -> SystemErr {
        println!("Queued -> SystemErr");
        SystemErr {}
    }
}

impl Starting {
    pub fn on_step(self, _: Step) -> Running {
        println!("Starting -> Running");
        Running {}
    }

    pub fn on_cancelled(self, _: Cancelled) -> Cancelled {
        println!("Starting -> Cancelled");
        Cancelled {}
    }

    pub fn on_system_err(self, _: SystemErr) -> SystemErr {
        println!("Starting -> SystemErr");
        SystemErr {}
    }
}

impl Running {
    pub fn on_step(self, _: Step) -> Running {
        println!("Running -> Running");
        Running {}
    }

    pub fn on_finishing(self, _: Finishing) -> Finishing {
        println!("Running -> Finishing");
        Finishing {}
    }

    pub fn on_cancelled(self, _: Cancelled) -> Cancelled {
        println!("Running -> Cancelled");
        Cancelled {}
    }

    pub fn on_system_err(self, _: SystemErr) -> SystemErr {
        println!("Running -> SystemErr");
        SystemErr {}
    }
}

impl Finishing {
    pub fn on_fail(self, _: Fail) -> Fail {
        println!("Finishing -> Fail");
        Fail {}
    }

    pub fn on_done(self, _: Done) -> Done {
        println!("Finishing -> Done");
        Done {}
    }
}

#[derive(Clone, Debug, PartialEq)]
pub struct BuildContext {
    pub org: String,
    pub repo_name: String,
    pub branch: String,
    pub id: Option<i32>,
    pub hash: Option<String>,
    pub user_envs: Option<Vec<String>>,
    pub job_trigger: JobTrigger,
    pub queue_time: Option<NaiveDateTime>,
    pub start_time: Option<NaiveDateTime>,
    pub _build_config: Option<OrbitalConfig>,
    pub _repo_uri: Option<GitUrl>,
    _state: BuildState,
}

impl BuildContext {
    pub fn new() -> Self {
        BuildContext {
            org: "".to_string(),
            repo_name: "".to_string(),
            branch: "".to_string(),
            id: None,
            hash: None,
            user_envs: None,
            job_trigger: JobTrigger::Manual,
            queue_time: None,
            start_time: None,
            _build_config: None,
            _repo_uri: None,
            _state: BuildState::queued(),
        }
    }

    pub fn add_org(mut self, org: String) -> BuildContext {
        self.org = org;
        self
    }

    pub fn add_repo_uri(mut self, repo_uri: String) -> Result<BuildContext> {
        let repo_uri_parsed =
            git_info::git_remote_url_parse(repo_uri.as_ref()).expect("Could not parse repo uri");

        self.repo_name = repo_uri_parsed.name.clone();

        self._repo_uri = Some(repo_uri_parsed);

        Ok(self)
    }

    pub fn add_repo_name(mut self, repo_name: String) -> BuildContext {
        self.repo_name = repo_name;
        self
    }

    pub fn add_branch(mut self, branch: String) -> BuildContext {
        self.branch = branch;
        self
    }

    pub fn add_id(mut self, id: i32) -> BuildContext {
        self.id = Some(id);
        self
    }

    pub fn add_hash(mut self, hash: String) -> BuildContext {
        self.hash = Some(hash);
        self
    }

    pub fn add_triggered_by(mut self, trigger: JobTrigger) -> BuildContext {
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
            &self.repo_name,
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
            &self.repo_name,
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

    pub fn state(self) -> BuildState {
        self._state
    }

    // Change state only once then return
    // TODO: This needs to be changed to accept a channel to stream to
    pub fn step(self) -> Result<BuildContext> {
        let mut current_step = self.clone();

        // Check for termination conditions

        // Connect to database. Query for the repo
        let pg_conn = postgres::client::establish_connection();

        // Check if cancelled
        match postgres::client::is_build_canceled(
            &pg_conn,
            &current_step.org,
            &current_step.repo_name,
            &current_step.hash.clone().unwrap_or_default(),
            &current_step.branch,
            current_step.id.clone().unwrap(),
        ) {
            Ok(true) => {
                info!("Build was cancelled");
                current_step._state = current_step.clone().state().on_cancelled(Cancelled {});

                // TODO: Update database

                return Ok(current_step);
            }
            Ok(false) => {
                info!("Build was not cancelled");
            }
            _ => {
                error!("Error checking for build cancellation");
                current_step._state = current_step.clone().state().on_system_err(SystemErr {});

                // TODO: Update database

                return Ok(current_step);
            }
        };

        //next._state = match next.clone().state() == BuildState::finishing() {
        //    true => {
        //        // If there were failures during the run, then move to failed, otherwise move to done

        //        next.clone().state().on_done(Done {})
        //    } // TODO: Manage the cleanup. Diff between pass/fail.
        //    false => {
        //        // If there are more commands to run, then step, otherwise move to finishing
        //        next.clone().state().on_step(Step)
        //    }
        //};

        let next_step = match current_step.clone().state() {
            BuildState::Queued(_) => {
                // Get secrets for cloning
                let mut next_step = current_step.clone();
                next_step._state = next_step._state.clone().on_step(Step {});
                next_step
            }
            BuildState::Starting(_) => {
                // Clone code
                // Validate orb.yml
                // Set a start time
                // Initialize stage name and step index
                let mut next_step = current_step.clone();
                next_step._state = next_step._state.clone().on_step(Step {});
                next_step
            }
            BuildState::Running(_) => {
                // Run command per step index
                // If it is a new stage, create the metadata

                // Note exit code?

                // Mark next command to run
                // If this was the last command in a stage, mark the end time

                // If the

                let mut next_step = current_step.clone();
                // Run this step once to prove loopback works
                next_step._state = next_step._state.clone().on_step(Step {});

                // DEBUG! Stepping this to Finish state immediately
                next_step._state = next_step._state.clone().on_finishing(Finishing {});
                next_step
            }
            BuildState::Finishing(_) => {
                // Set the end time for the build

                let mut next_step = current_step.clone();
                next_step._state = next_step._state.clone().on_done(Done {});
                next_step
            }
            _ => current_step.clone(),
        };

        Ok(next_step)
    }
}
