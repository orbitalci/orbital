//#[macro_use]
extern crate structopt;

extern crate clap;

//use clap::arg_enum;
use structopt::StructOpt;

// This is for autocompletion
//use structopt::clap::Shell;

use subcommand;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Command {
    /// Trigger build for registered repo
    Build(subcommand::build_subcmd::SubOption),
    /// Manage credentials
    Creds(subcommand::creds::CredType),
    Init(subcommand::init::SubOption),
    Cancel(subcommand::cancel::SubOption),
    Logs(subcommand::logs::SubOption),
    Poll(subcommand::poll::SubOption),
    Repo(subcommand::repo::SubOption),
    Summary(subcommand::summary::SubOption),
    Validate(subcommand::validate::SubOption),
    Version(subcommand::version::SubOption),
    Watch(subcommand::watch::SubOption),
}

// TODO: Split the top-level options into a struct that we can pass to to the command handlers
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
        Command::Build(a) => subcommand::build_subcmd::subcommand_handler(a),
        Command::Creds(a) => subcommand::creds::subcommand_handler(a),
        Command::Init(a) => subcommand::init::subcommand_handler(a),
        Command::Cancel(a) => subcommand::cancel::subcommand_handler(a),
        Command::Logs(a)  => subcommand::logs::subcommand_handler(a),
        Command::Poll(a) => subcommand::poll::subcommand_handler(a),
        Command::Repo(a) => subcommand::repo::subcommand_handler(a),
        Command::Summary(a) => subcommand::summary::subcommand_handler(a),
        Command::Validate(a) => subcommand::validate::subcommand_handler(a),
        Command::Version(a) => subcommand::version::subcommand_handler(a),
        Command::Watch(a) => subcommand::watch::subcommand_handler(a),
    }

    //println!("Full matches: {:?}", matches);
}