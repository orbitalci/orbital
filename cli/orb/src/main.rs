use color_eyre::eyre::Result;
use structopt::StructOpt;

extern crate clap;

use subcommand::{self, Subcommand, SubcommandContext};

use serde::{Deserialize, Serialize};

use log::debug;
use std::env;
use std::fs::OpenOptions;

#[derive(Clone, Debug, PartialEq, Serialize, Deserialize)]
struct OrbCliConfigFile {
    debug: Option<bool>,
    default_org: Option<String>,
}

/// Parse command line input, and route into one of the subcommand handlers along with global options
#[tokio::main]
async fn main() -> Result<()> {
    // Read in a file-based config if it exists
    let homedir = env::var("HOME").expect("HOME env var not set");
    let _config_file = match OpenOptions::new()
        .read(true)
        .open(format!("{}/.orbrc", homedir))
    {
        Ok(conf) => {
            println!("Config file found");
            let parsed_config: OrbCliConfigFile = serde_yaml::from_reader(conf).unwrap();

            if let Some(debug_flag) = parsed_config.debug {
                if env::var_os("RUST_LOG").is_none() && debug_flag {
                    let debug_modules = vec![
                        "subcommand",
                        "orbital_services",
                        "orbital_agent",
                        "orbital_database",
                        "git_meta",
                        "hashicorp_stack",
                        "git_url_parse",
                    ];

                    env::set_var("RUST_LOG", debug_modules.join(","))
                }
            }

            if let Some(default_org) = parsed_config.default_org {
                if env::var_os("ORB_DEFAULT_ORG").is_none() {
                    env::set_var("ORB_DEFAULT_ORG", default_org)
                }
            }
        }
        Err(_) => debug!(
            "Could not open config file. Either doesn't exist or permissions prevent reading."
        ),
    };

    env_logger::init();

    // If it exists, load it up
    // set env vars as appropriate so they get used by the subcommands
    // If env vars are already set, then _don't_ override them

    let parsed = SubcommandContext::from_args();

    // Pass to the subcommand handlers
    match parsed.subcommand {
        Subcommand::Build(sub_option) => {
            subcommand::build_cmd::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Cancel(sub_option) => {
            subcommand::cancel::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Logs(sub_option) => {
            subcommand::logs::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Org(sub_option) => {
            subcommand::org::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Repo(sub_option) => {
            subcommand::repo::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Secret(sub_option) => {
            subcommand::secret::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Summary(sub_option) => {
            subcommand::summary::subcommand_handler(parsed.global_option, sub_option).await
        }
        Subcommand::Server(sub_command) => {
            subcommand::server::subcommand_handler(parsed.global_option, sub_command).await
        }
        Subcommand::Developer(sub_command) => {
            subcommand::developer::subcommand_handler(parsed.global_option, sub_command).await
        }
    }
}
