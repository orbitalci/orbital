/// This is named build_subcmd.rs bc we can't use build.rs due to overlapping with `cargo` features.

extern crate clap;

use structopt::StructOpt;

use std::env;


use git_meta::git_info::{self,GitSshRemote};
use git2::Repository;
use itertools::structs::Format;

//ocelot build -acct-repo <acct>/<repo> -hash <git_hash> -branch <branch> [-latest]

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Build provided account/repo. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Use provided local branch. Default to current active branch
    #[structopt(long)]
    branch: Option<String>,
    /// Build provided commit hash. Otherwise, default to HEAD commit of active branch
    #[structopt(long)]
    hash: Option<String>,
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
    
}

// The goal of this command
// If we pass a commit hash alone, we assume the current branch.
//      If no, then we might end up in a detached HEAD? Is there a way I can walk a commit to a branch?
//      If yes, then we should pass back a remote ref to the branch+commit

// If we pass a local branch alone, we should resolve the branch to a remote ref HEAD

// Passing both the branch and commit should resolve to that specific remote ref

// TODO: This should return a Result for the questionmark operator
// Handle the command line control flow
pub fn subcommand_handler(args: &SubOption) {

    // Parse the following information from the local repo before calling the backend
    //
    // git provider Account name
    // Repo (something.git)
    // Provider (like bitbucket or github account)
    // Remote Branch
    // Target commit, or HEAD of the remote branch if not specified
    // Later: Env vars

    // Assume current directory for now
    let path_to_repo = args.path.clone()
                        .unwrap_or(
                            env::current_dir()
                            .unwrap().to_str()
                            .unwrap()
                            .to_string()
                        );

    println!("Path to repo: {:?}", path_to_repo);


    // Get the git info from the path
    let git_info = git_info::get_git_info_from_path(&path_to_repo, &args.branch, &args.hash);
    println!("Git info: {:?}",git_info);




    //let local_repo_ref = git_info::get_local_ref_from_path(path_to_repo.to_str().unwrap());


    //let remote_url = git_info::git_remote_from_ref(&local_repo_ref);
    //println!("{:?}", remote_url);

    //// Going to assume we're only supporting SSH remotes for now
    //let remote = git_info::git_remote_url_parse(&remote_url);
    //println!("{:?}", remote);

    //// Determine the specifics about what to build
    //// Get branch
    //// Get commit hash or choose HEAD
    //let reference = git_info::get_local_ref(path_to_repo.to_str().unwrap(), &args.branch, &args.hash);
    //let branch = git_info::get_working_branch(&local_repo_ref, &None);

    //println!("{:?}", branch.unwrap().name());
}
