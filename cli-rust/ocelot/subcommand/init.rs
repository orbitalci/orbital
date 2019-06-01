extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    #[structopt(name = "Machine tag", long)]
    machine_tag: Option<bool>,

    #[structopt(name = "Slack", long)]
    slack: Option<bool>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling init");
}