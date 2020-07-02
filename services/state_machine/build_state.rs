use machine::{machine, transitions};
use orbital_headers::orbital_types::JobState as ProtoJobState;

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

impl From<BuildState> for ProtoJobState {
    fn from(build_state: BuildState) -> Self {
        match build_state {
            BuildState::Queued(Queued {}) => ProtoJobState::Queued,
            BuildState::Starting(Starting {}) => ProtoJobState::Starting,
            BuildState::Running(Running {}) => ProtoJobState::Running,
            BuildState::Finishing(Finishing {}) => ProtoJobState::Finishing,
            BuildState::Done(Done {}) => ProtoJobState::Done,
            BuildState::Cancelled(Cancelled {}) => ProtoJobState::Cancelled,
            BuildState::Fail(Fail {}) => ProtoJobState::Failed,
            BuildState::SystemErr(SystemErr {}) => ProtoJobState::SystemErr,
            BuildState::Error => ProtoJobState::Unknown,
        }
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
    pub fn on_step(self, _: Step) -> Starting {
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
