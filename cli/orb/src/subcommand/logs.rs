extern crate structopt;
use structopt::StructOpt;

use crate::{GlobalOption, SubcommandError};

use orbital_headers::build_meta::{client::BuildServiceClient, BuildTarget};

use crate::ORB_DEFAULT_URI;
use tonic::Request;
//use futures::stream;

/// Local options for customizing logs request
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long)]
    path: Option<String>,
}

/// Generates request for logs
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    _local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    let mut client = BuildServiceClient::connect(format!("http://{}", ORB_DEFAULT_URI)).await?;


    // Need to figure out how to handle the streaming response

    //let request = Request::new(BuildTarget {
    //    org: "org_name_goes_here".into(),
    //    ..Default::default()
    //});


    //let mut stream = client
    //.build_logs(Request::new(request))
    //.await?
    //.into_inner();

    //while let Some(response) = stream.message().await? {
    //    println!("RESPONSE = {:?}", response);
    //}


    Ok(())
}
