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

    let commit = get_target_commit(&local_repo, &Some(working_branch.clone()), commit_id)?;

    let commit_id = format!("{}", &commit.id());

    let commit_msg = commit.clone().message().unwrap_or_default().to_string();

    Ok(GitCommitContext {
        commit_id: commit_id,
        branch: working_branch,
        message: commit_msg.to_string(),
        git_url: GitUrl::parse(&remote_url)?,
    })
}

// FIXME: Should not assume remote is origin. This will cause issue in some dev workflows
/// Returns the remote url after opening and validating repo from the local path
pub fn git_remote_from_path(path: &Path) -> Result<String> {
    // Theory: After we get the repo, we can derive the remote name from the branch

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
fn git_remote_from_repo(local_repo: &Repository) -> Result<String> {
    let remote_url: String = local_repo
        .find_remote("origin")?
        .url()
        .expect("Unable to extract repo url from remote")
        .chars()
        .collect();
    Ok(remote_url)
}

/// Returns `GitUrl` after parsing input git url
pub fn git_remote_url_parse(remote_url: &str) -> Result<GitUrl> {
    GitUrl::parse(remote_url)
}

/// Return the `git2::Branch` struct for a local repo (as opposed to a remote repo)
/// If `local_branch` is not provided, we'll select the current active branch, based on HEAD
fn get_working_branch<'repo>(
    r: &'repo Repository,
    local_branch: &Option<String>,
) -> Result<Branch<'repo>> {
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

            // Convert git2::Error to anyhow::Error
            match r.find_branch(
                local_branch
                    .name()?
                    .expect("Unable to return local branch name"),
                BranchType::Local,
            ) {
                Ok(b) => Ok(b),
                Err(e) => Err(e.into()),
            }
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
) -> Result<Commit<'repo>> {
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
    use git_url_parse::Scheme;

    #[test]
    fn parse_github_ssh_url() -> Result<()> {
        let gh_url_parsed = git_remote_url_parse("git@github.com:orbitalci/orbital.git")?;

        let expected_parsed = GitUrl {
            host: Some("github.com".to_string()),
            name: "orbital".to_string(),
            owner: Some("orbitalci".to_string()),
            organization: None,
            fullname: "orbitalci/orbital".to_string(),
            scheme: Scheme::Ssh,
            user: Some("git".to_string()),
            token: None,
            port: None,
            path: "orbitalci/orbital.git".to_string(),
            git_suffix: true,
            scheme_prefix: false,
        };

        assert_eq!(gh_url_parsed, expected_parsed);
        Ok(())
    }

    #[test]
    fn parse_bitbucket_ssh_url() -> Result<()> {
        let bb_url_parsed = git_remote_url_parse("git@bitbucket.com:orbitalci/orbital.git")?;

        let expected_parsed = GitUrl {
            host: Some("bitbucket.com".to_string()),
            name: "orbital".to_string(),
            owner: Some("orbitalci".to_string()),
            organization: None,
            fullname: "orbitalci/orbital".to_string(),
            scheme: Scheme::Ssh,
            user: Some("git".to_string()),
            token: None,
            port: None,
            path: "orbitalci/orbital.git".to_string(),
            git_suffix: true,
            scheme_prefix: false,
        };

        assert_eq!(bb_url_parsed, expected_parsed);
        Ok(())
    }

    #[test]
    fn parse_azure_ssh_url() -> Result<()> {
        let az_url_parsed =
            git_remote_url_parse("git@ssh.dev.azure.com:v3/organization/project/repo")?;

        let expected_parsed = GitUrl {
            host: Some("ssh.dev.azure.com".to_string()),
            name: "repo".to_string(),
            owner: Some("organization".to_string()),
            organization: Some("organization".to_string()),
            fullname: "v3/organization/project/repo".to_string(),
            scheme: Scheme::Ssh,
            user: Some("git".to_string()),
            token: None,
            port: None,
            path: "v3/organization/project/repo".to_string(),
            git_suffix: false,
            scheme_prefix: false,
        };

        assert_ne!(az_url_parsed, expected_parsed);
        Ok(())
    }
}
