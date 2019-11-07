extern crate structopt;
use structopt::StructOpt;

use serde::{Deserialize, Serialize};

use std::fs::File;
use std::io::prelude::*;
use std::path::Path;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(name = "Machine tag", long)]
    machine_tag: Option<bool>,

    #[structopt(name = "Slack", long)]
    slack: Option<bool>,
}

// TODO: Move all of these config structs to another crate. Perhaps the ocelot_api
#[derive(Debug, PartialEq, Serialize, Deserialize, Default)]
struct OcelotNotifySlackBlock {
    channel: String,
    identifiers: String,
    on: Vec<String>,
}

#[derive(Debug, PartialEq, Serialize, Deserialize, Default)]
struct OcelotNotify {
    slack: OcelotNotifySlackBlock,
}

#[derive(Debug, PartialEq, Serialize, Deserialize, Default)]
struct OcelotConfigStageTrigger {
    branches: Vec<String>,
}

#[derive(Debug, PartialEq, Serialize, Deserialize, Default)]
struct OcelotConfigStage {
    name: String,
    trigger: OcelotConfigStageTrigger,
    script: Vec<String>,
    env: Vec<String>,
}

#[derive(Debug, PartialEq, Serialize, Deserialize, Default)]
struct OcelotConfig {
    version: String,
    build_tool: String,
    notify: OcelotNotify,
    branches: Vec<String>,
    env: Vec<String>,
    stages: Vec<OcelotConfigStage>,
}

// Handle the command line control flow
pub fn subcommand_handler(_args: SubOption) {
    println!("Placeholder for handling init");

    let branch_trigger = OcelotConfigStageTrigger {
        branches: ["master".to_string()].to_vec(),
    };
    let ocelot_stage = OcelotConfigStage {
        name: "Test".to_string(),
        trigger: branch_trigger,
        env: vec![],
        script: ["echo hello world".to_string()].to_vec(),
    };

    let ocelot_notify = OcelotNotify {
        slack: OcelotNotifySlackBlock {
            channel: "".to_string(),
            identifiers: "".to_string(),
            on: vec!["FAIL".to_string()],
        },
    };

    let ocelot_config = OcelotConfig {
        version: 1.to_string(),
        build_tool: "docker".to_string(),
        notify: ocelot_notify,
        branches: vec!["master".to_string()],
        env: vec![],
        stages: vec![ocelot_stage],
    };

    let mut config_default: OcelotConfig = Default::default();
    config_default.stages = Vec::new();

    println!("Config: {:?}", serde_yaml::to_string(&ocelot_config));

    match Path::new("./ocelot.yml").exists() {
        true => println!("ocelot.yml exists in path. Skipping"),
        false => {
            println!("Create ocelot.yml");
            let mut file = File::create("ocelot.yml").unwrap();
            let _ = file.write_all(serde_yaml::to_string(&ocelot_config).unwrap().as_bytes());
        }
    }
}
