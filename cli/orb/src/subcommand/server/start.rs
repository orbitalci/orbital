use color_eyre::eyre::Result;
use structopt::StructOpt;

use crate::subcommand::GlobalOption;

use orbital_headers::{
    build_meta::build_service_server::BuildServiceServer,
    code::code_service_server::CodeServiceServer,
    notify::notify_service_server::NotifyServiceServer,
    organization::organization_service_server::OrganizationServiceServer,
    secret::secret_service_server::SecretServiceServer,
};
use orbital_services::OrbitalApi;

use orbital_services::ORB_DEFAULT_URI;

use log::info;
use std::env;
use std::path::PathBuf;

// For the service router
use futures::future::{self, Either, TryFutureExt};
use http::version::Version;
use hyper::{service::make_service_fn, Server as HyperServer};
use std::convert::Infallible;
use std::{
    pin::Pin,
    task::{Context, Poll},
};
use tower::Service;
use warp::Filter;

type Error = Box<dyn std::error::Error + Send + Sync + 'static>;

/// Local options for starting build service
#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct SubcommandOption {
    /// Path to local repo. Defaults to current working directory
    #[structopt(long, parse(from_os_str), env = "PWD")]
    path: PathBuf,

    #[structopt(long)]
    debug: bool,

    // The polling frequency, in seconds
    #[structopt(long, default_value = "60")]
    poll_freq: u8,
}

/// Binds a *currently hardcoded* address and starts all services on mutliplexed gRPC server
pub async fn subcommand_handler(
    _global_option: GlobalOption,
    local_option: SubcommandOption,
) -> Result<()> {
    let addr = ORB_DEFAULT_URI.parse().unwrap();

    if local_option.debug {
        if env::var_os("RUST_LOG").is_none() {
            let debug_modules = vec![
                "subcommand::server",
                "orbital_services",
                "orbital_agent",
                "orbital_database",
                "git_meta",
                "hashicorp_stack",
                "git_url_parse",
                "git_event",
            ];

            env::set_var("RUST_LOG", debug_modules.join(","))
        }
    }

    let _ = env_logger::try_init();

    // Kick off thread for checking for new commits
    {
        info!("Starting new commit polling");
        crate::subcommand::server::poll::poll_for_new_commits(local_option.poll_freq).await;
    }

    info!("Starting single-node server");
    let warp = warp::service(warp::path("hello").map(|| "hello, world!"));

    HyperServer::bind(&addr)
        .serve(make_service_fn(move |_| {
            let mut tonic = tonic::transport::Server::builder()
                .add_service(BuildServiceServer::new(OrbitalApi::default()))
                .add_service(CodeServiceServer::new(OrbitalApi::default()))
                .add_service(NotifyServiceServer::new(OrbitalApi::default()))
                .add_service(OrganizationServiceServer::new(OrbitalApi::default()))
                .add_service(SecretServiceServer::new(OrbitalApi::default()))
                .into_service();

            let mut warp = warp.clone();

            future::ok::<_, Infallible>(tower::service_fn(
                move |req: hyper::Request<hyper::Body>| match req.version() {
                    Version::HTTP_11 | Version::HTTP_10 => Either::Left(
                        warp.call(req)
                            .map_ok(|res| res.map(EitherBody::Left))
                            .map_err(Error::from),
                    ),
                    Version::HTTP_2 => Either::Right(
                        tonic
                            .call(req)
                            .map_ok(|res| res.map(EitherBody::Right))
                            .map_err(Error::from),
                    ),
                    _ => unimplemented!(),
                },
            ))
        }))
        .await?;

    Ok(())
}

// From Tonic example: hyper_warp/server
enum EitherBody<A, B> {
    Left(A),
    Right(B),
}

impl<A, B> http_body::Body for EitherBody<A, B>
where
    A: http_body::Body + Send + Unpin,
    B: http_body::Body<Data = A::Data> + Send + Unpin,
    A::Error: Into<Error>,
    B::Error: Into<Error>,
{
    type Data = A::Data;
    type Error = Box<dyn std::error::Error + Send + Sync + 'static>;

    fn is_end_stream(&self) -> bool {
        match self {
            EitherBody::Left(b) => b.is_end_stream(),
            EitherBody::Right(b) => b.is_end_stream(),
        }
    }

    fn poll_data(
        self: Pin<&mut Self>,
        cx: &mut Context<'_>,
    ) -> Poll<Option<Result<Self::Data, Self::Error>>> {
        match self.get_mut() {
            EitherBody::Left(b) => Pin::new(b).poll_data(cx).map(map_option_err),
            EitherBody::Right(b) => Pin::new(b).poll_data(cx).map(map_option_err),
        }
    }

    fn poll_trailers(
        self: Pin<&mut Self>,
        cx: &mut Context<'_>,
    ) -> Poll<Result<Option<http::HeaderMap>, Self::Error>> {
        match self.get_mut() {
            EitherBody::Left(b) => Pin::new(b).poll_trailers(cx).map_err(Into::into),
            EitherBody::Right(b) => Pin::new(b).poll_trailers(cx).map_err(Into::into),
        }
    }
}

fn map_option_err<T, U: Into<Error>>(err: Option<Result<T, U>>) -> Option<Result<T, Error>> {
    err.map(|e| e.map_err(Into::into))
}
