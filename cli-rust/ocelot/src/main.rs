//#[macro_use]
extern crate structopt;

extern crate clap;

//use clap::arg_enum;
use structopt::StructOpt;

// This is for autocompletion
//use structopt::clap::Shell;

use subcommand;

#[derive(Debug, StructOpt, Copy, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    Add,
    Delete,
    List,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Command {
    /// Trigger build for registered repo
    Build(subcommand::build_subcmd::SubOption),
    /// Manage credentials
    Creds(subcommand::creds::CredType),
    Init,
    Kill,
    Logs,
    Poll(ResourceAction),
    Repos(ResourceAction),
    Status,
    Summary,
    Validate,
    Version,
    Watch,
}

// FIXME: Need to think about how to pass any of the options to the subcommands
#[derive(Debug, StructOpt)]
#[structopt(name = "ocelot")]
/// The OrbitalCI command line interface
pub struct ApplicationArguments {
    #[structopt(subcommand)]
    pub command: Command,
    #[structopt(long = "consul-addr")]
    /// http api address of Consul. Specified as URI with scheme (e.g http://127.0.0.1:8500)
    pub consul_addr: Option<String>,
    #[structopt(long = "vault-addr")]
    /// http api address of Vault. Specified as URI with scheme (e.g http://127.0.0.1:8200)
    pub vault_addr: Option<String>,
    #[structopt(long = "vault-token")]
    pub vault_token: Option<String>,

    #[structopt(long = "nsqd-addr")]
    pub nsqd_addr: Option<String>,
    #[structopt(long = "nsqlookupd-addr")]
    pub nsqlookupd_addr: Option<String>,

    #[structopt(long = "debug")]
    pub debug : bool,

}


// TODO: Can we define traits to keep a tighter contract for creds?
fn main() {
    // generate `bash` completions in "target" directory
    //ApplicationArguments::clap().gen_completions(env!("CARGO_PKG_NAME"), Shell::Bash, "target");   

    let matches = ApplicationArguments::from_args();

    // Do stuff with the optional args

    // Pass to the subcommand handlers
    match &matches.command {
        Command::Build(a) => {
            subcommand::build_subcmd::subcommand_handler(a);
        },
        Command::Creds(a) => {
            subcommand::creds::subcommand_handler(a);
        },
        Command::Init => {},
        Command::Kill => {},
        Command::Logs  => {},
        Command::Poll(_a) => {},
        Command::Repos(_a) => {},
        Command::Status => {},
        Command::Summary => {},
        Command::Validate => {},
        Command::Version => {},
        Command::Watch => {},
    }

    //println!("Full matches: {:?}", matches);
}