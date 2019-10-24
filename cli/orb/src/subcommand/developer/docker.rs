extern crate structopt;
use std::str::FromStr;
use structopt::StructOpt;

use shiplift::{
    builder::ContainerFilter, tty::StreamType, ContainerListOptions, ContainerOptions, Docker,
    ExecContainerOptions, PullOptions,
};
use tokio;
use tokio::prelude::{Future, Stream};

use crate::{GlobalOption, SubcommandError};

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Docker image
    #[structopt(long)]
    image: Option<String>,

    /// Pull, Create
    action: Action,
}

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum Action {
    Pull,
    Create,
}

impl FromStr for Action {
    type Err = String;
    fn from_str(action: &str) -> Result<Self, Self::Err> {
        match action {
            "pull" => Ok(Action::Pull),
            "create" => Ok(Action::Create),
            _ => Err("Invalid action".to_string()),
        }
    }
}

fn container_pull(image: Option<String>) -> Result<(), ()> {
    let docker = Docker::new();

    let img = match image {
        Some(i) => i,
        None => "alpine:latest".to_string(),
    };

    println!("Pulling image: {}", img);

    let img_pull = docker
        .images()
        .pull(&PullOptions::builder().image(img.clone()).build())
        .for_each(|output| {
            println!("{:?}", output);
            Ok(())
        })
        .map_err(|e| eprintln!("Error: {}", e));
    Ok(tokio::run(img_pull))
    //Ok(())
}

pub fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<(), SubcommandError> {
    //container_pull(local_option.image);

    match local_option.action {
        Action::Pull => {
            container_pull(local_option.image);
        }
        Action::Create => println!("Placeholder. Create container."),
    }
    Ok(())
}
