use orbital_headers::orbital_types::JobTrigger;
//use postgres::schema::JobTrigger;

#[derive(Debug)]
struct BuildStateQueued;
#[derive(Debug)]
struct BuildStateStarting;
#[derive(Debug)]
struct BuildStateRunning;
#[derive(Debug)]
struct BuildStateCanceled;
#[derive(Debug)]
struct BuildStateSystemErr;
#[derive(Debug)]
struct BuildStateFailed;
#[derive(Debug)]
struct BuildStateDone;

#[derive(Debug)]
struct BuildContext<S> {
    id: Option<u32>,
    hash: Option<String>,
    job_trigger: JobTrigger,
    _state: S,
}

impl BuildContext<BuildStateQueued> {
    pub fn new() -> Self {
        BuildContext {
            id: None,
            hash: None,
            job_trigger: JobTrigger::Manual,
            _state: BuildStateQueued,
        }
    }

    pub fn id(mut self, id: u32) -> BuildContext<BuildStateQueued> {
        self.id = Some(id);
        self
    }

    pub fn hash(mut self, hash: String) -> BuildContext<BuildStateQueued> {
        self.hash = Some(hash);
        self
    }
}
