// Dig into a filepath and harvest info about the local branch

// Get as much info about the remote branch as well

use git2::{Branch, BranchType, Commit, ObjectType, Repository};

use log::debug;

pub fn get_local_repo_from_path(path: &str) -> Result<Repository, git2::Error> {
    Repository::open(path)
}

// Clean this up with stronger types later
#[derive(Debug, Default)]
pub struct GitCommitContext {
    pub provider: String,
    pub branch: String,
    pub id: String,
    pub account: String,
    pub repo: String,
}

pub fn get_git_info_from_path(
    path: &str,
    branch: &Option<String>,
    commit_id: &Option<String>,
) -> Result<GitCommitContext, git2::Error> {
    // Our return struct
    let mut commit = GitCommitContext::default();

    // First we open the repository and get the remote_url and parse it into components
    let local_repo = get_local_repo_from_path(path)?;
    let remote_info = git_remote_url_parse(&git_remote_from_repo(&local_repo)?);

    // TODO: Do this in two stages, we could support passing a remote branch, and then fall back to a local branch
    // Assuming we are passed a local branch from remote "origin", or nothing.
    // Let's make sure it resolves to a remote branch
    let working_branch = get_working_branch(&local_repo, branch)?
        .name()?
        .expect("Unable to extract branch name")
        .to_string();

    let working_commit = get_target_commit(&local_repo, &Some(working_branch.clone()), commit_id);

    commit.provider = remote_info.provider;
    commit.account = remote_info.account;
    commit.repo = remote_info.repo;
    commit.branch = working_branch;

    commit.id = format!("{}", working_commit?.id());

    Ok(commit)
}

pub fn git_remote_from_path(path: &str) -> Result<String, git2::Error> {
    let r = get_local_repo_from_path(path)?;
    let remote_url: String = r
        .find_remote("origin")?
        .url()
        .expect("Unable to extract repo url from remote")
        .chars()
        .collect();
    Ok(remote_url)
}

pub fn git_remote_from_repo(local_repo: &Repository) -> Result<String, git2::Error> {
    let remote_url: String = local_repo
        .find_remote("origin")?
        .url()
        .expect("Unable to extract repo url from remote")
        .chars()
        .collect();
    Ok(remote_url)
}

#[derive(Debug)]
pub struct GitSshRemote {
    user: String,
    provider: String,
    account: String,
    repo: String,
}

// FIXME: This parser fails to select the correct account and repo names on azure ssh repo uris. Off by one
// Example
// ssh: git@ssh.dev.azure.com:v3/organization/project/repo
// http: https://organization@dev.azure.com/organization/project/_git/repo
pub fn git_remote_url_parse(remote_url: &str) -> GitSshRemote {
    // TODO: We will want to see if we can parse w/ Url, since git repos might use HTTPS
    //let http_url = Url::parse(remote_url);
    // If we get Err(RelativeUrlWithoutBase) then we should pick apart the remote url
    //println!("{:?}",http_url);

    // Splitting on colon first will split
    // user@provider.tld:account/repo.git
    let split_first_stage = remote_url.split(":").collect::<Vec<&str>>();

    let user_provider = split_first_stage[0].split("@").collect::<Vec<&str>>();
    let acct_repo = split_first_stage[1].split("/").collect::<Vec<&str>>();

    let repo_parsed = acct_repo[1].to_string();
    let repo_parsed = repo_parsed.split(".git").collect::<Vec<&str>>();

    GitSshRemote {
        user: user_provider[0].to_string(),
        provider: user_provider[1].to_string(),
        account: acct_repo[0].to_string(),
        repo: repo_parsed[0].to_string(),
    }
}

// Assuming we are probably dealing with someone's local branch and it needs to be mapped to a remote branch
// If we haven't passed one in, use the active branch
pub fn get_working_branch<'repo>(
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

pub fn is_commit_in_branch<'repo>(r: &'repo Repository, commit: &Commit, branch: &Branch) -> bool {
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

pub fn get_target_commit<'repo>(
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
