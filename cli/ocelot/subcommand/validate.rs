extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// File path to yaml containing build configuration. Defaults to looking in current directory
    #[structopt(name = "Artifact repo yaml", short = "f", long = "file")]
    file_path: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(_args: SubOption) {
    println!("Placeholder for handling validation");
}