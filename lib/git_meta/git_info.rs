// Dig into a filepath and harvest info about the local branch

// Get as much info about the remote branch as well

use anyhow::Result;
use git2::{Branch, BranchType, Commit, ObjectType, Repository};
use git_url_parse::GitUrl;
use log::debug;
use std::path::Path;

use super::GitCommitContext;

/// Returns a `git2::Repository` from a given repo directory path
fn get_local_repo_from_path(path: &Path) -> Result<Repository, git2::Error> {
    Repository::open(path.as_os_str())
}

/// Returns a `GitCommitContext` after parsing metadata from a repo
/// If branch is not provided, current checked out branch will be used
/// If commit id is not provided, the HEAD of the branch will be used
pub fn get_git_info_from_path(
    path: &Path,
    branch: &Option<String>,
    commit_id: &Option<String>,
) -> Result<GitCommitContext> {
    // First we open the repository and get the remote_url and parse it into components
    let local_repo = get_local_repo_from_path(path)?;
    let remote_url = git_remote_from_repo(&local_repo)?;

    // TODO: Do this in two stages, we could support passing a remote branch, and then fall back to a local branch
    // Assuming we are passed a local branch from remote "origin", or nothing.
    // Let's make sure it resolves to a remote branch
    let working_branch = get_working_branch(&local_repo, branch)?
        .name()?
        .expect("Unable to extract branch name")
        .to_string();

    let working_commit = format!(
        "{}",
        get_target_commit(&local_repo, &Some(working_branch.clone()), commit_id)?.id()
    );

    Ok(GitCommitContext {
        commit_id: working_commit,
        branch: working_branch,
        git_url: GitUrl::parse(&remote_url)?,
    })
}

// FIXME: Should not assume remote is origin. This will cause issue in some dev workflows
/// Returns the remote url after opening and validating repo from the local path
pub fn git_remote_from_path(path: &Path) -> Result<String, git2::Error> {
    let r = get_local_repo_from_path(path)?;
    let remote_url: String = r
        .find_remote("origin")?
        .url()
        .expect("Unable to extract repo url from remote")
        .chars()
        .collect();
    Ok(remote_url)
}

/// Returns the remote url from the `git2::Repository` struct
fn git_remote_from_repo(local_repo: &Repository) -> Result<String, git2::Error> {
    let remote_url: String = local_repo
        .find_remote("origin")?
        .url()
        .expect("Unable to extract repo url from remote")
        .chars()
        .collect();
    Ok(remote_url)
}

// FIXME: This parser fails to select the correct account and repo names on azure ssh repo uris. Off by one
// Example
// ssh: git@ssh.dev.azure.com:v3/organization/project/repo
// http: https://organization@dev.azure.com/organization/project/_git/repo
/// Return a `GitSshRemote` after parsing a remote url from a git repo
pub fn git_remote_url_parse(remote_url: &str) -> Result<GitUrl> {
    GitUrl::parse(remote_url)
}

/// Return the `git2::Branch` struct for a local repo (as opposed to a remote repo)
/// If `local_branch` is not provided, we'll select the current active branch, based on HEAD
fn get_working_branch<'repo>(
    r: &'repo Repository,
    local_branch: &Option<String>,
) -> Result<Branch<'repo>, git2::Error> {
    // It is likely that a user branch will not contain the remote

    match local_branch {
        Some(branch) => {
            //println!("User passed branch: {:?}", branch);
            let b = r.find_branch(&branch, BranchType::Local)?;
            debug!("Returning given branch: {:?}", &b.name());
            Ok(b)
        }
        None => {
            // Getting the HEAD of the current
            let head = r.head();
            //let commit = head.unwrap().peel_to_commit();
            //println!("{:?}", commit);

            // Find the current local branch...
            let local_branch = Branch::wrap(head?);

            debug!("Returning HEAD branch: {:?}", local_branch.name()?);
            r.find_branch(
                local_branch
                    .name()?
                    .expect("Unable to return local branch name"),
                BranchType::Local,
            )

            //r.find_branch(&"master", BranchType::Local)
        }
    }
}

/// Returns a `bool` if the `git2::Commit` is a descendent of the `git2::Branch`
fn is_commit_in_branch<'repo>(r: &'repo Repository, commit: &Commit, branch: &Branch) -> bool {
    let branch_head = branch.get().peel_to_commit();

    if branch_head.is_err() {
        return false;
    }

    let branch_head = branch_head.expect("Unable to extract branch HEAD commit");
    if branch_head.id() == commit.id() {
        return true;
    }

    // We get here if we're not working with HEAD commits, and we gotta dig deeper

    let check_commit_in_branch = r.graph_descendant_of(branch_head.id(), commit.id());
    //println!("is {:?} a decendent of {:?}: {:?}", &commit.id(), &branch_head.id(), is_commit_in_branch);

    if check_commit_in_branch.is_err() {
        return false;
    }

    check_commit_in_branch.expect("Unable to determine if commit exists within branch")
}

// TODO: Verify if commit is not in branch, that we'll end up in detached HEAD
/// Return a `git2::Commit` that refers to the commit object requested for building
/// If commit id is not provided, then we'll use the HEAD commit of whatever branch is active or provided
fn get_target_commit<'repo>(
    r: &'repo Repository,
    branch: &Option<String>,
    commit_id: &Option<String>,
) -> Result<Commit<'repo>, git2::Error> {
    let working_branch = get_working_branch(r, branch)?;
    let working_ref = working_branch.into_reference();

    match commit_id {
        Some(id) => {
            debug!("Commit provided. Using {}", id);
            let oid = git2::Oid::from_str(id)?;

            let obj = r.find_object(oid, ObjectType::from_str("commit"))?;
            let commit = obj
                .into_commit()
                .expect("Unable to convert commit id into commit object");

            let _ = is_commit_in_branch(r, &commit, &Branch::wrap(working_ref));

            Ok(commit)
        }
        // We want the HEAD of the working branch
        None => {
            debug!("No commit provided. Using HEAD commit");
            let commit = working_ref
                .peel_to_commit()
                .expect("Unable to retrieve HEAD commit object from working branch");

            let _ = is_commit_in_branch(r, &commit, &Branch::wrap(working_ref));

            Ok(commit)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_github_ssh_url() -> Result<(), String> {
        let gh_url_parsed = git_remote_url_parse("git@github.com:level11consulting/orbitalci.git");

        let expected_parsed = GitSshRemote {
            user: "git".to_string(),
            provider: "github.com".to_string(),
            account: "level11consulting".to_string(),
            repo: "orbitalci".to_string(),
        };

        assert_eq!(gh_url_parsed, expected_parsed);
        Ok(())
    }

    #[test]
    fn parse_bitbucket_ssh_url() -> Result<(), String> {
        let bb_url_parsed =
            git_remote_url_parse("git@bitbucket.com:level11consulting/orbitalci.git");

        let expected_parsed = GitSshRemote {
            user: "git".to_string(),
            provider: "bitbucket.com".to_string(),
            account: "level11consulting".to_string(),
            repo: "orbitalci".to_string(),
        };

        assert_eq!(bb_url_parsed, expected_parsed);
        Ok(())
    }

    // This is a negative test. https://github.com/level11consulting/orbitalci/issues/228
    // We need to keep track of this so we can accomodate for parser changes
    #[test]
    fn parse_azure_ssh_url() -> Result<(), String> {
        let az_url_parsed =
            git_remote_url_parse("git@ssh.dev.azure.com:v3/organization/project/repo");

        let expected_parsed = GitSshRemote {
            user: "git".to_string(),
            provider: "ssh.dev.azure.com".to_string(),
            account: "organization".to_string(),
            repo: "repo".to_string(),
        };

        assert_ne!(az_url_parsed, expected_parsed);
        Ok(())
    }
}
