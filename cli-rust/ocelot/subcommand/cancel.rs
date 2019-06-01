//--render-tag -- add a machineTag field, instead of image
//--notify -- Add a section for slack

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Git commit hash
    #[structopt(long)]
    hash: Option<String>,

    /// Build ID
    #[structopt(name = "build id", long)]
    build_id: Option<u32>,

    /// Branch 
    #[structopt(name = "branch", long)]
    branch: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for handling build cancellation");
}