/// This is named build_subcmd.rs bc we can't use build.rs due to overlapping with `cargo` features.

extern crate clap;

use structopt::StructOpt;

use std::env;


use git_meta::git_info;
use git2::Repository;
use itertools::structs::Format;

//ocelot build -acct-repo <acct>/<repo> -hash <git_hash> -branch <branch> [-latest]

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Build provided account/repo. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Use provided branch. Default to current active branch
    #[structopt(long)]
    branch : Option<String>,
    /// Build provided commit hash. Otherwise, default to HEAD commit of active branch
    #[structopt(long)]
    hash : Option<String>,
}

// Let's define the build options here
pub fn build() {
    println!("Hello, world!");
}
// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {
    println!("Placeholder for running build");

    // Assume current directory for now
    let path_to_repo = env::current_dir().unwrap();

    // Get build information

    // Parse remote from url
    // We want this information before calling the backend
    //
    // Account, Repo
    // Provider
    // Remote Branch
    // Hash, or HEAD if not specified
    // Later: Env vars


    let remote_url = git_info::git_remote_from_path(path_to_repo.to_str().unwrap());
    println!("{:?}", remote_url);



    // Get commit hash or choose HEAD
}