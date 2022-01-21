use shiplift::{
    tty::TtyChunk, ContainerOptions, Docker, ExecContainerOptions, LogsOptions, PullOptions,
};

use futures::StreamExt;

use color_eyre::eyre::{eyre, Result};
use tracing::debug;
use std::time::Duration;

use serde_json::value::Value;
use tokio::sync::mpsc;

#[derive(Debug, Default, Clone)]
pub struct OrbitalContainerSpec<'a> {
    pub name: Option<String>,
    pub image: String,
    pub command: Vec<&'a str>,
    pub env_vars: Option<Vec<&'a str>>,
    pub volumes: Option<Vec<&'a str>>,
    pub timeout: Option<Duration>,
}

/// Returns a String to work around shiplift behavior that is different from the docker cli
/// If we give shiplift an image w/o a tag, it'll download all the tags. Usually the intended behavior is to only pull latest
/// ```
/// let injected_latest = container_builder::docker::image_tag_sanitizer("docker").unwrap();
/// assert_eq!("docker:latest", injected_latest);
///
/// let tag_provided = container_builder::docker::image_tag_sanitizer("alpine:3").unwrap();
/// assert_eq!("alpine:3", tag_provided);
/// ```
pub fn image_tag_sanitizer(image: &str) -> Result<String> {
    let split = &image.split(':').collect::<Vec<_>>();

    match split.len() {
        1 => {
            return {
                debug!("Image tag was not provided. Assuming {}:latest", &image);
                Ok(format!("{}:latest", image))
            }
        }
        2 => Ok(image.to_string()),
        _ => Err(eyre!("Failed to clean docker image tag")),
    }
}

/// Connect to the docker engine and pull the provided image
/// if no tag is provided with the image, ":latest" tag will be assumed
pub async fn container_pull<S: AsRef<str>>(image: S) -> Result<mpsc::UnboundedReceiver<Value>> {
    let (tx, rx) = mpsc::unbounded_channel();

    let docker = Docker::new();

    let img = image_tag_sanitizer(image.as_ref())?;

    debug!("Pulling image: {}", img);

    let mut stream = docker
        .images()
        .pull(&PullOptions::builder().image(img.clone()).build());

    while let Some(pull_result) = stream.next().await {
        match pull_result {
            Ok(output) => {
                debug!("{:?}", output);

                let _ = match tx.send(output) {
                    Ok(_) => Ok(()),
                    Err(_) => Err(()),
                };
            }
            Err(e) => eprintln!("Error: {}", e),
        }
    }

    Ok(rx)
}

/// Connect to the docker engine and create a container
/// Currently assumes that source code gets mounted in container's /orbital-work directory
/// Returns the id of the container that is created
pub async fn container_create(container_spec: OrbitalContainerSpec<'_>) -> Result<String, ()> {
    let docker = Docker::new();

    let env_vec: Vec<&str> = container_spec.env_vars.unwrap_or_default();
    debug!("Adding env vars: {:?}", env_vec);
    let volume_vec: Vec<&str> = container_spec.volumes.unwrap_or_default();
    debug!("Adding volume mounts: {:?}", volume_vec);

    // TODO: Need a naming convention

    let container_spec = match &container_spec.name {
        Some(container_name) => ContainerOptions::builder(&container_spec.image)
            .name(container_name)
            .attach_stdout(true)
            .attach_stderr(true)
            .working_dir(super::ORBITAL_CONTAINER_WORKDIR)
            .env(env_vec)
            .volumes(volume_vec)
            .cmd(container_spec.command)
            .build(),
        None => ContainerOptions::builder(&container_spec.image)
            .attach_stdout(true)
            .attach_stderr(true)
            .working_dir(super::ORBITAL_CONTAINER_WORKDIR)
            .env(env_vec)
            .volumes(volume_vec)
            .cmd(container_spec.command)
            .build(),
    };

    match docker.containers().create(&container_spec).await {
        Ok(info) => Ok(info.id),
        Err(_e) => Err(()),
    }
}

/// Connect to the docker engine and start a created container with a given `container_id`
pub async fn container_start(container_id: &str) -> Result<()> {
    let docker = Docker::new();

    debug!("Starting the container");

    match docker
        .containers()
        .get(String::from(container_id))
        .start()
        .await
    {
        Ok(_) => Ok(()),
        Err(_) => Err(eyre!("Could not start container")),
    }
}

/// Connect to the docker engine and stop a running container with a given `container_id`
pub async fn container_stop(container_id: &str) -> Result<()> {
    let docker = Docker::new();
    match docker
        .containers()
        .get(String::from(container_id))
        .stop(None)
        .await
    {
        Ok(_) => Ok(()),
        Err(_) => Err(eyre!("Could not stop container")),
    }
}

/// Connect to the docker engine and execute commands in a running container with a given `container_id`
pub async fn container_exec<S: AsRef<str>>(
    container_id: S,
    command: Vec<&str>,
) -> Result<mpsc::UnboundedReceiver<String>> {
    let (tx, rx) = mpsc::unbounded_channel();
    let docker = Docker::new();

    //println!("Executing commands in the container");
    let options = ExecContainerOptions::builder()
        .cmd(command)
        .attach_stdout(true)
        .attach_stderr(true)
        .build();

    // send output to channel
    let mut exec_stream = docker
        .containers()
        .get(container_id.as_ref())
        .exec(&options);

    while let Some(exec_result) = exec_stream.next().await {
        match exec_result {
            Ok(chunk) => match chunk {
                TtyChunk::StdOut(bytes) => {
                    tx.send(std::str::from_utf8(&bytes).unwrap().to_string())
                        .unwrap();
                    print!("{}", std::str::from_utf8(&bytes).unwrap())
                }
                TtyChunk::StdErr(bytes) => {
                    tx.send(std::str::from_utf8(&bytes).unwrap().to_string())
                        .unwrap();
                    eprintln!("{}", std::str::from_utf8(&bytes).unwrap())
                }
                TtyChunk::StdIn(_) => unreachable!(),
            },
            Err(e) => eprintln!("Error: {}", e),
        }
    }
    Ok(rx)
}

pub async fn container_logs<S: AsRef<str>>(
    container_id: S,
) -> Result<mpsc::UnboundedReceiver<String>> {
    let (tx, rx) = mpsc::unbounded_channel();

    let docker = Docker::new();

    // send output to channel
    let mut logs_stream = docker
        .containers()
        .get(container_id.as_ref())
        .logs(&LogsOptions::builder().stdout(true).stderr(true).build());

    while let Some(log_result) = logs_stream.next().await {
        match log_result {
            Ok(chunk) => match chunk {
                TtyChunk::StdOut(bytes) => {
                    tx.send(std::str::from_utf8(&bytes).unwrap().to_string())
                        .unwrap();
                    print!("{}", std::str::from_utf8(&bytes).unwrap())
                }
                TtyChunk::StdErr(bytes) => {
                    tx.send(std::str::from_utf8(&bytes).unwrap().to_string())
                        .unwrap();
                    eprintln!("{}", std::str::from_utf8(&bytes).unwrap())
                }
                TtyChunk::StdIn(_) => unreachable!(),
            },
            Err(_e) => eprintln!("Error"),
        }
    }

    Ok(rx)
}
