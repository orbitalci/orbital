use shiplift::{tty::StreamType, ContainerOptions, Docker, ExecContainerOptions, PullOptions};

use tokio;
use tokio::prelude::{Future, Stream};

use anyhow::{anyhow, Result};
use log::{debug, error};
use std::sync::mpsc::channel;

/// Returns a String to work around shiplift behavior that is different from the docker cli
/// If we give shiplift an image w/o a tag, it'll download all the tags. Usually the intended behavior is to only pull latest
/// ```
/// let injected_latest = container_builder::docker::image_tag_sanitizer("docker").unwrap();
/// assert_eq!("docker:latest", injected_latest);
///
/// let tag_provided = container_builder::docker::image_tag_sanitizer("alpine:3").unwrap();
/// assert_eq!("alpine:3", tag_provided);
/// ```
#[doc(test(no_crate_inject))]
pub fn image_tag_sanitizer(image: &str) -> Result<String> {
    let split = &image.split(":").collect::<Vec<_>>();

    match split.len() {
        1 => {
            return {
                debug!("Image tag was not provided. Assuming {}:latest", &image);
                Ok(format!("{}:latest", image))
            }
        }
        2 => return Ok(image.to_string()),
        //_ => return Err(anyhow::Error::msg("Failed to clean docker image tag")),
        _ => return Err(anyhow!("Failed to clean docker image tag")),
    }
}

/// Connect to the docker engine and pull the provided image
/// if no tag is provided with the image, ":latest" tag will be assumed
pub fn container_pull(image: &str) -> Result<()> {
    let docker = Docker::new();

    let img = image_tag_sanitizer(image)?;

    debug!("Pulling image: {}", img);

    let img_pull = docker
        .images()
        .pull(&PullOptions::builder().image(img.clone()).build())
        .for_each(|output| {
            debug!("{:?}", output);
            Ok(())
        })
        .map_err(|e| eprintln!("Error: {}", e));
    Ok(tokio::run(img_pull))
}

/// Connect to the docker engine and create a container
/// Currently assumes that source code gets mounted in container's /orbital-work directory
/// Returns the id of the container that is created
pub fn container_create(
    image: &str,
    command: Vec<&str>,
    envs: Option<Vec<&str>>,
    vols: Option<Vec<&str>>,
) -> Result<String, ()> {
    let docker = Docker::new();

    let env_vec: Vec<&str> = envs.unwrap_or_default();
    debug!("Adding env vars: {:?}", env_vec);
    let volume_vec: Vec<&str> = vols.unwrap_or_default();
    debug!("Adding volume mounts: {:?}", volume_vec);

    // TODO: Need a naming convention
    let container_spec = ContainerOptions::builder(image)
        //.name("test-container-name")
        .attach_stdout(true)
        .attach_stderr(true)
        .working_dir(super::ORBITAL_CONTAINER_WORKDIR)
        .env(env_vec)
        .volumes(volume_vec)
        .cmd(command)
        .build();

    let new_container = docker
        .containers()
        .create(&container_spec)
        .map(|info| {
            debug!("{:?}", info);
            info
        })
        .map_err(|e| {
            error!("Error: {}", e);
            e
        });

    let mut container_runtime = tokio::runtime::Runtime::new().expect("Unable to create a runtime");

    // Wait for the container to be created so we can get its container id
    let container = match container_runtime.block_on(new_container) {
        Ok(runtime) => runtime,
        Err(_) => return Err(()),
    };

    Ok(container.id)
}

/// Connect to the docker engine and start a created container with a given `container_id`
pub fn container_start(container_id: &str) -> Result<()> {
    let docker = Docker::new();

    debug!("Starting the container");

    let start_container = docker
        .containers()
        .get(&container_id)
        .start()
        .map(|info| {
            println!("{:?}", info);
            info
        })
        .map_err(|e| eprintln!("Error: {}", e));
    tokio::run(start_container);

    Ok(())
}

/// Connect to the docker engine and stop a running container with a given `container_id`
pub fn container_stop(container_id: &str) -> Result<()> {
    let docker = Docker::new();
    let stop_container = docker
        .containers()
        .get(&container_id)
        .stop(None)
        .map(|info| {
            println!("{:?}", info);
            info
        })
        .map_err(|e| eprintln!("Error: {}", e));
    tokio::run(stop_container);
    Ok(())
}

// TODO: This will need a mechanism to stream logs out to a client
/// Connect to the docker engine and execute commands in a running container with a given `container_id`
pub fn container_exec(container_id: &str, command: Vec<&str>) -> Result<Vec<String>> {
    let docker = Docker::new();

    println!("{:?}", command);
    // FYI: This might not work on MacOS hosts until https://github.com/softprops/shiplift/issues/155 is fixed
    println!("Executing commands in the container");
    let options = ExecContainerOptions::builder()
        .cmd(command)
        .attach_stdout(true)
        .attach_stderr(true)
        .build();


    // open a channel
    let (tx, rx) = channel();

    // send output to channel
    let exec_container = docker
        .containers()
        .get(&container_id)
        .exec(&options)
        .for_each(move |chunk| {
            match chunk.stream_type {
                StreamType::StdOut => {
                    tx.send(chunk.as_string_lossy()).unwrap();
                    print!("{}", chunk.as_string_lossy())
                },
                StreamType::StdErr => {
                    tx.send(chunk.as_string_lossy()).unwrap();
                    eprintln!("{}", chunk.as_string_lossy());
                },
                StreamType::StdIn => unreachable!(),
            }
            Ok(())
        })
        .map_err(|e| eprintln!("Error: {}", e));

    tokio::run(exec_container);


    let exec_output : Vec<String> = rx.into_iter().collect();

    Ok(exec_output)
}
