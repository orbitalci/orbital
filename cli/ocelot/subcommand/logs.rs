//ocelot logs --hash <git_hash>
// add --build-id
// add --acct-repo

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Build ID
    #[structopt(name = "build id", long)]
    build_id: Option<u32>,
    /// Retrieve logs for account/repo. Otherwise try to auto-detect from current working directory
    #[structopt(long)]
    acct_repo: Option<String>,
    /// Retrieve logs for the provided branch. Without build-id or hash, will default to latest commit in branch
    #[structopt(long)]
    branch: Option<String>,
    /// Retrieve logs for the provided commit hash. Otherwise, default to latest build
    #[structopt(long)]
    hash: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(_args: SubOption) {
    println!("Placeholder for handling logs");

    //let uri = ocelot_api::client_util::get_client_uri();
    //let dst = Destination::try_from_uri(uri.clone()).unwrap();

    //let connector = util::Connector::new(HttpConnector::new(4));
    //let settings = client::Builder::new().http2_only(true).clone();
    //let mut make_client = client::Connect::with_builder(connector, settings);
}
