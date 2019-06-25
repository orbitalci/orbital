extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Check if there are newer updates available
    #[structopt(long)]
    check_update: Option<bool>,
}

// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    const VERSION: Option<&'static str> = option_env!("CARGO_PKG_VERSION");

    println!("Ocelot - Rust client v{}", VERSION.unwrap_or("unknown"));
}
