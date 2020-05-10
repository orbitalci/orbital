use structopt::StructOpt;

use crate::{repo::SubcommandOption, GlobalOption};

use orbital_headers::code::{code_service_client::CodeServiceClient, GitRepoAddRequest};
use orbital_headers::orbital_types::SecretType;
use orbital_services::ORB_DEFAULT_URI;
use tonic::Request;

use git_meta::git_info;
use git_url_parse::Protocol;
use log::{debug, info};
use std::fs::File;
use std::io::prelude::*;
use std::path::PathBuf;

use anyhow::Result;
use orbital_database::postgres::repo::Repo;
use prettytable::{cell, format, row, Table};

#[derive(Debug, StructOpt, Clone)]
#[structopt(rename_all = "kebab_case")]
pub struct ActionOption {
    /// Repo path
    #[structopt(parse(from_os_str), env = "PWD")]
    path: PathBuf,

    /// Name of Orbital org
    #[structopt(long, env = "ORB_DEFAULT_ORG")]
    org: Option<String>,

    /// Use flag if repo is public
    #[structopt(long)]
    public: bool,

    // TODO: We're only supporting ssh key auth from the client at the moment.
    /// Path to private key
    #[structopt(
        long,
        parse(from_os_str),
        required_unless("public"),
        required_unless("password")
    )]
    private_key: Option<PathBuf>,

    /// Password for private repo. Mutually exclusive w/ private key
    #[structopt(long, required_unless("public"), required_unless("private-key"))]
    password: Option<String>,

    /// Username for private repo
    #[structopt(long, short = "u")]
    username: Option<String>,

    /// Skip checking branch clone before adding
    #[structopt(long)]
    skip_check: bool,

    /// Check clone with provided branch instead of master
    #[structopt(long)]
    alt_branch: Option<String>,
}

pub async fn action_handler(
    _global_option: GlobalOption,
    _subcommand_option: SubcommandOption,
    action_option: ActionOption,
) -> Result<()> {
    // Read git repo info from local filesystem.
    // TODO: support adding a repo from a uri
    let repo_info =
        match git_info::get_git_info_from_path(&action_option.path.as_path(), &None, &None) {
            Ok(info) => info,
            Err(_e) => panic!("Unable to parse path for git repo info"),
        };

    info!("Adding repo: {:?}", &repo_info);

    // TODO: Need to update the git repo parser to split out a username
    let request = match &action_option.public {
        true => {

            if repo_info.clone().git_url.protocol == Protocol::Ssh {
                panic!("Repo auth is not public")
            };

            info!("Repo is public");
            // TODO
            // If the repo proto is ssh and we're specifying that this is public
            // just error out now instead of making remote call.

            Request::new(GitRepoAddRequest {
                org: action_option
                    .org
                    .clone()
                    .expect("Please provide an org with request"),
                secret_type: SecretType::Unspecified.into(),
                git_provider: repo_info.git_url.host.unwrap(),
                name: repo_info.git_url.name,
                uri: repo_info.git_url.href,
                user: repo_info.git_url.user.unwrap_or_default(),
                alt_check_branch: action_option.alt_branch.unwrap_or_default(),
                skip_check: action_option.skip_check,
                ..Default::default()
            })
        }
        false => {
            info!("Repo is private");

            match action_option.private_key {
                Some(p) => {
                    info!("Repo auth with private key");

                    // Read in private key into memory
                    let mut file = File::open(p.to_str().expect("No secret filepath given"))?;
                    let mut contents = String::new();
                    file.read_to_string(&mut contents)?;

                    Request::new(GitRepoAddRequest {
                        org: action_option
                            .org
                            .clone()
                            .expect("Please provide an org with request"),
                        secret_type: SecretType::SshKey.into(),
                        auth_data: contents,
                        git_provider: repo_info.git_url.host.unwrap(),
                        name: repo_info.git_url.name,
                        uri: repo_info.git_url.href,
                        user: repo_info.git_url.user.unwrap(),
                        alt_check_branch: action_option.alt_branch.unwrap_or_default(),
                        skip_check: action_option.skip_check,
                    })
                }
                None => {
                    info!("Repo auth with basic auth");

                    Request::new(GitRepoAddRequest {
                        org: action_option
                            .org
                            .clone()
                            .expect("Please provide an org with request"),
                        secret_type: SecretType::BasicAuth.into(),
                        auth_data: action_option.password.expect("No password provided"),
                        git_provider: repo_info.git_url.host.unwrap(),
                        name: repo_info.git_url.name,
                        uri: repo_info.git_url.href,
                        user: action_option.username.expect("No username provided"),
                        alt_check_branch: action_option.alt_branch.unwrap_or_default(),
                        skip_check: action_option.skip_check,
                    })
                }
            }
        }
    };

    debug!("Request for git repo add: {:?}", &request);

    let mut client = CodeServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;
    let response = client.git_repo_add(request).await;

    match response {
        Err(_e) => {
            eprintln!("Error adding Repo");
            Ok(())
        }
        Ok(o) => {
            let repo_proto = o.into_inner();

            debug!("RESPONSE = {:?}", &repo_proto);

            // By default, format the response into a table
            let mut table = Table::new();
            table.set_format(*format::consts::FORMAT_NO_BORDER_LINE_SEPARATOR);

            // Print the header row
            table.set_titles(row![
                bc =>
                "Org Name",
                "Repo Name",
                "Uri",
                "Secret Name",
                "Build Enabled",
                "Notify Enabled",
                "Next build index"
            ]);

            let repo = Repo::from(repo_proto.clone());

            table.add_row(row![
                action_option.org.unwrap(),
                repo.name,
                repo.uri,
                repo_proto.auth_data,
                &format!("{:?}", repo.build_active_state),
                &format!("{:?}", repo.notify_active_state),
                repo.next_build_index
            ]);

            // Print the table to stdout
            table.printstd();
            Ok(())
        }
    }
}
