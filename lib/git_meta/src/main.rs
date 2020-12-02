use git_meta::git_info;
use std::path::Path;

fn main() {
    //let old_commit = "a23cf309d85f27116cc748bfd925aec9b2902e24";
    let old_commit = "b12fe2fd1233b398d0d523b670085ba8e124ae44";

    // Assumptions. We already have the repo cloned.
    // We cannot make that assumption in prod.

    // open the repo
    let latest_commit = git_info::get_git_info_from_path(Path::new("../.."), &None, &None).unwrap();
    println!("Latest commit: {:?}", latest_commit.commit_id);

    match latest_commit.commit_id == old_commit {
        true => println!("No new build"),
        false => println!("New build"),
    };

    // If the old commit is different than latest_commit, then we assume that new commits have been made. We should queue.

    //let latest_commit_hash = get_latest_commit(repo, branch_filter, credentials);
    //check if latest_commit_hash is newer than last known commit
    //if newer then queue a new build
    //sleep 30s
}
