use structopt::StructOpt;

extern crate clap;

use subcommand::{self, GlobalOption, Subcommand, SubcommandContext, SubcommandError};

/// Parse command line input, and route into one of the subcommand handlers along with global options
#[tokio::main]
async fn main() -> Result<(), SubcommandError> {
    env_logger::init();

    let parsed = SubcommandContext::from_args();

    // Pass to the subcommand handlers
    match parsed.subcommand {
        Subcommand::Build(sub_option) => {
            subcommand::build_cmd::subcommand_handler(parsed.global_option, sub_option).await

            // we can default to following the logs using the metadata info that we'll need to capture
            //
            //subcommand::logs::subcommand_handler(parsed.global_option, sub_option).await
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
        Subcommand::Operator(sub_command) => {
            subcommand::operator::subcommand_handler(parsed.global_option, sub_command).await
        }
        Subcommand::Developer(sub_command) => {
            subcommand::developer::subcommand_handler(parsed.global_option, sub_command).await
        }
    }
}
