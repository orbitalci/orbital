use structopt::StructOpt;

extern crate clap;
use std::io;

use subcommand::{self, GlobalOption, Subcommand, SubcommandContext, SubcommandError};

fn main() -> Result<(), SubcommandError> {
    env_logger::init();

    let parsed = SubcommandContext::from_args();

    // Pass to the subcommand handlers
    match parsed.subcommand {
        Subcommand::Build(sub_option) => {
            subcommand::build_cmd::subcommand_handler(parsed.global_option, sub_option)
        }
        Subcommand::Cancel => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Logs => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Org => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Repo => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Poll => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Secret => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Summary => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Operator(sub_command) => {
            subcommand::operator::subcommand_handler(parsed.global_option, sub_command)
        }
        Subcommand::Developer(sub_command) => {
            subcommand::developer::subcommand_handler(parsed.global_option, sub_command)
        }
        Subcommand::Version => Err(SubcommandError::new("Not yet implemented")),
        Subcommand::Completion(shell) => {
            GlobalOption::clap().gen_completions_to(
                env!("CARGO_PKG_NAME"),
                shell.into(),
                &mut io::stdout(),
            );
            Ok(())
        }
    }

}
