use crate::postgres::org::Org;
use crate::postgres::schema::{repo, ActiveState, GitHostType};
use crate::postgres::secret::Secret;

use orbital_headers::code::GitRepoEntry;
//use orbital_headers::secret::SecretEntry;

use git_meta::git_info;

#[derive(Insertable, Debug, PartialEq, Associations, AsChangeset)]
#[belongs_to(Org)]
#[belongs_to(Secret)]
#[table_name = "repo"]
pub struct NewRepo {
    pub org_id: i32,
    pub name: String,
    pub uri: String,
    pub git_host_type: GitHostType,
    pub secret_id: Option<i32>,
    pub build_active_state: ActiveState,
    pub notify_active_state: ActiveState,
    pub next_build_index: i32,
}

impl Default for NewRepo {
    fn default() -> Self {
        NewRepo {
            org_id: 0,
            name: "".into(),
            uri: "".into(),
            git_host_type: GitHostType::Generic,
            secret_id: None,
            build_active_state: ActiveState::Enabled,
            notify_active_state: ActiveState::Enabled,
            next_build_index: 0,
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
    pub git_host_type: GitHostType,
    pub secret_id: Option<i32>,
    pub build_active_state: ActiveState,
    pub notify_active_state: ActiveState,
    pub next_build_index: i32,
}

impl Default for Repo {
    fn default() -> Self {
        Repo {
            id: 0,
            org_id: 0,
            name: "".into(),
            uri: "".into(),
            git_host_type: GitHostType::Generic,
            secret_id: None,
            build_active_state: ActiveState::Enabled,
            notify_active_state: ActiveState::Enabled,
            next_build_index: 0,
        }
    }
}

// FIXME: Org should be a string, but right now we only have the postgres org id
impl From<Repo> for GitRepoEntry {
    fn from(repo: Repo) -> Self {
        let git_uri_parsed = git_info::git_remote_url_parse(&repo.uri.clone()).unwrap();

        //
        GitRepoEntry {
            org: repo.org_id.to_string(), // FIXME: We should have the org name
            git_provider: git_uri_parsed.host.unwrap(),
            name: git_uri_parsed.name,
            user: git_uri_parsed.user.unwrap(),
            //uri: git_uri_parsed.repo, // FIXME: THis is an inconsistency
            //secret_type:
            //auth_data:
            build: repo.build_active_state.into(),
            notify: repo.notify_active_state.into(),
            next_build_index: repo.next_build_index,
            ..Default::default()
        }
    }
}

// FIXME: This does not correctly set the org id
impl From<GitRepoEntry> for Repo {
    fn from(git_repo_entry: GitRepoEntry) -> Self {
        let git_uri_parsed = git_info::git_remote_url_parse(&git_repo_entry.uri.clone()).unwrap();

        Repo {
            id: 0,
            //org_id: git_repo_entry.org,
            org_id: 0,
            name: git_repo_entry.name.clone(),
            uri: git_uri_parsed.href,
            //git_host_type: git_repo_entry.,
            //secret_id: Option<i32>,
            build_active_state: git_repo_entry.build.into(),
            notify_active_state: git_repo_entry.notify.into(),
            next_build_index: git_repo_entry.next_build_index,
            ..Default::default()
        }
    }
}
