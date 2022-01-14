use crate::orbital_database::postgres::org::Org;
use crate::orbital_database::postgres::schema::{repo, ActiveState, GitHostType};
use crate::orbital_database::postgres::secret::Secret;
use serde_json::json;

use crate::orbital_headers::code::{
    GitRepoEntry, GitRepoRemoteBranchHead, GitRepoRemoteBranchHeadList,
};
//use orbital_headers::secret::SecretEntry;

use git_url_parse::GitUrl;
use log::warn;

#[derive(Insertable, Debug, PartialEq, Associations, AsChangeset)]
#[belongs_to(Org)]
#[belongs_to(Secret)]
#[table_name = "repo"]
pub struct NewRepo {
    pub org_id: i32,
    pub name: String,
    pub uri: String,
    pub canonical_branch: String,
    pub git_host_type: GitHostType,
    pub secret_id: Option<i32>,
    pub build_active_state: ActiveState,
    pub notify_active_state: ActiveState,
    pub next_build_index: i32,
    pub remote_branch_heads: serde_json::Value,
}

impl Default for NewRepo {
    fn default() -> Self {
        NewRepo {
            org_id: 0,
            name: "".into(),
            uri: "".into(),
            canonical_branch: "".into(),
            git_host_type: GitHostType::Generic,
            secret_id: None,
            build_active_state: ActiveState::Enabled,
            notify_active_state: ActiveState::Enabled,
            next_build_index: 1,
            remote_branch_heads: json!([]),
        }
    }
}

#[derive(Clone, Debug, Identifiable, Queryable, Associations, QueryableByName)]
#[belongs_to(Org)]
#[belongs_to(Secret)]
#[table_name = "repo"]
pub struct Repo {
    pub id: i32,
    pub org_id: i32,
    pub name: String,
    pub uri: String,
    pub canonical_branch: String,
    pub git_host_type: GitHostType,
    pub secret_id: Option<i32>,
    pub build_active_state: ActiveState,
    pub notify_active_state: ActiveState,
    pub next_build_index: i32,
    pub remote_branch_heads: serde_json::Value,
}

impl Default for Repo {
    fn default() -> Self {
        Repo {
            id: 0,
            org_id: 0,
            name: "".into(),
            uri: "".into(),
            canonical_branch: "".into(),
            git_host_type: GitHostType::Generic,
            secret_id: None,
            build_active_state: ActiveState::Enabled,
            notify_active_state: ActiveState::Enabled,
            next_build_index: 1,
            remote_branch_heads: json!([]),
        }
    }
}

// FIXME: Org should be a string, but right now we only have the postgres org id
impl From<Repo> for GitRepoEntry {
    fn from(repo: Repo) -> Self {
        let git_uri_parsed = GitUrl::parse(&repo.uri).unwrap();

        //
        GitRepoEntry {
            org: repo.org_id.to_string(), // FIXME: We should have the org name
            git_provider: git_uri_parsed.host.unwrap(),
            name: git_uri_parsed.name,
            user: git_uri_parsed.user.unwrap(),
            uri: repo.uri,
            canonical_branch: repo.canonical_branch,
            //secret_type
            //auth_data:
            build: repo.build_active_state.into(),
            notify: repo.notify_active_state.into(),
            next_build_index: repo.next_build_index,
            remote_branch_heads: {
                match repo.remote_branch_heads {
                    serde_json::Value::Null => None,
                    serde_json::Value::Object(map_value) => {
                        let mut git_branches: Vec<GitRepoRemoteBranchHead> = Vec::new();

                        for (k, v) in map_value {
                            let branch = GitRepoRemoteBranchHead {
                                branch: k,
                                commit: v.to_string(),
                            };

                            git_branches.push(branch);
                        }
                        Some(GitRepoRemoteBranchHeadList {
                            remote_branch_heads: git_branches,
                        })
                    }
                    _ => {
                        warn!("There was a serde Value other than an Object. Dropping value. ");
                        None
                    }
                }
            },
            ..Default::default()
        }
    }
}

// FIXME: This does not correctly set the org id
impl From<GitRepoEntry> for Repo {
    fn from(git_repo_entry: GitRepoEntry) -> Self {
        let git_uri_parsed = GitUrl::parse(&git_repo_entry.uri).unwrap();

        Repo {
            id: 0,
            //org_id: git_repo_entry.org,
            org_id: 0,
            name: git_repo_entry.name.clone(),
            uri: format!("{}", git_uri_parsed),
            canonical_branch: git_repo_entry.canonical_branch,
            //git_host_type: git_repo_entry.,
            //secret_id: Option<i32>,
            build_active_state: git_repo_entry.build.into(),
            notify_active_state: git_repo_entry.notify.into(),
            next_build_index: git_repo_entry.next_build_index,
            remote_branch_heads: {
                match git_repo_entry.remote_branch_heads {
                    // TODO: Unpack this shit
                    Some(_branches) => {
                        json!([])
                    }
                    None => json!([]),
                }
            },
            ..Default::default()
        }
    }
}
