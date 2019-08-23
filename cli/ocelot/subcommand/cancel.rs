//--render-tag -- add a machineTag field, instead of image
//--notify -- Add a section for slack

extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubOption {
    /// Git commit hash
    #[structopt(long)]
    hash: Option<String>,

    /// Build ID
    #[structopt(name = "build id", long)]
    build_id: Option<u32>,

    /// Branch
    #[structopt(name = "branch", long)]
    branch: Option<String>,
}

// Handle the command line control flow
pub fn subcommand_handler(_args: SubOption) {
    println!("Placeholder for handling build cancellation");

    //let uri = ocelot_api::client_util::get_client_uri();
    //let dst = Destination::try_from_uri(uri.clone()).unwrap();

    //let connector = util::Connector::new(HttpConnector::new(4));
    //let settings = client::Builder::new().http2_only(true).clone();
    //let mut make_client = client::Connect::with_builder(connector, settings);
}
