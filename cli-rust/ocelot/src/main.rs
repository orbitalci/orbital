//#[macro_use]
extern crate structopt;

extern crate clap;

//use clap::arg_enum;
use structopt::StructOpt;

// This is for autocompletion
//use structopt::clap::Shell;

use subcommand::{self,build_subcmd,action};

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum CredType {
    Apple(action::ResourceAction),
    Env(action::ResourceAction),
    Helmrepo(action::ResourceAction),
    K8s(action::ResourceAction),
    Notify(action::ResourceAction),
    Repo(action::ResourceAction),
    Ssh(action::ResourceAction),
    Vcs(subcommand::creds::vcs::VcsOptions),
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Command {
    Build(build_subcmd::BuildOptions),
    Creds(CredType),
    Init,
    Kill,
    Logs,
    Poll(action::ResourceAction),
    Repos(action::ResourceAction),
    Status,
    Summary,
    Validate,
    Version,
    Watch,
}

#[derive(Debug, StructOpt)]
#[structopt(name = "orb")]
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


fn main() {
    // generate `bash` completions in "target" directory
    //ApplicationArguments::clap().gen_completions(env!("CARGO_PKG_NAME"), Shell::Bash, "target");   

    let matches = ApplicationArguments::from_args();

    // Do stuff with the optional args

    // Pass to the subcommand handlers
    match &matches.command {
        Command::Build(_) => {
            build_subcmd::build();
        },
        Command::Creds(_n) => {},
        Command::Init => {},
        Command::Kill => {},
        Command::Logs  => {},
        Command::Poll(_n) => {},
        Command::Repos(_n) => {},
        Command::Status => {},
        Command::Summary => {},
        Command::Validate => {},
        Command::Version => {},
        Command::Watch => {},
    }

    println!("Full matches: {:?}", matches);
}