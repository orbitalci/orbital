use crate::orbital_headers::build_meta::{
    build_service_server::BuildService, BuildLogResponse, BuildMetadata, BuildRecord, BuildStage,
    BuildSummaryRequest, BuildSummaryResponse, BuildTarget,
};

use crate::orbital_database::postgres;
use crate::orbital_database::postgres::build_summary::NewBuildSummary;
use crate::orbital_database::postgres::schema::JobState;
use crate::orbital_headers;
use crate::orbital_headers::orbital_types::JobState as ProtoJobState;
use chrono::{NaiveDateTime, Utc};
use mktemp::Temp;
use tonic::{Code, Request, Response, Status};

use tokio::sync::mpsc;

use crate::orbital_services::orbital_agent::build_engine;

use super::state_machine::build_context::BuildContext;
use super::state_machine::build_state::BuildState;
use super::OrbitalApi;
use crate::orbital_utils::orbital_agent;

use log::{debug, error, info};

use futures_util::FutureExt;

use tokio::runtime::Runtime;

// TODO: If this bails anytime before the end, we need to attempt some cleanup
/// Implementation of protobuf derived `BuildService` trait
#[tonic::async_trait]
impl BuildService for OrbitalApi {
    /// Start a build in a container. (Stay focused.)
    ///
    type BuildStartStream = tokio_stream::wrappers::ReceiverStream<Result<BuildRecord, Status>>;
    async fn build_start(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<Self::BuildStartStream>, Status> {
        //println!("DEBUG: {:?}", response);

        // Git clone for provider ( uri, branch, commit )
        let unwrapped_request = request.into_inner();

        info!("build request: {:?}", &unwrapped_request.git_repo);
        debug!("build request details: {:?}", &unwrapped_request);

        let (client_tx, client_rx) = mpsc::channel(1);

        let (build_tx, mut build_rx): (
            mpsc::UnboundedSender<String>,
            mpsc::UnboundedReceiver<String>,
        ) = mpsc::unbounded_channel();

        tokio::spawn(async move {
            let git_clone_dir = Temp::new_dir().expect("Unable to create dir for git clone");
            let private_key =
                Temp::new_file().expect("Unable to create temp private key file for git clone");

            let mut cur_build = BuildContext::new()
                .add_org(unwrapped_request.org.to_string())
                .add_repo_uri(unwrapped_request.clone().remote_uri)
                .expect("Could not parse repo uri")
                .add_branch(unwrapped_request.branch.to_string())
                .add_hash(unwrapped_request.commit_hash.to_string())
                .add_triggered_by(unwrapped_request.trigger.into())
                .add_working_dir(git_clone_dir.to_path_buf())
                .add_private_key(private_key.to_path_buf())
                .queue()
                .expect("There was a problem queuing the build");

            if !unwrapped_request.config.clone().is_empty() {
                cur_build = cur_build
                    .clone()
                    .add_build_config_from_string(unwrapped_request.config.clone())
                    .expect("Build config failed to parse");
            }

            'build_loop: loop {
                if (cur_build.clone().state() == BuildState::done())
                    | (cur_build.clone().state() == BuildState::cancelled())
                    | (cur_build.clone().state() == BuildState::fail())
                    | (cur_build.clone().state() == BuildState::system_err())
                {
                    debug!("Exiting build loop - {:?}", cur_build.clone().state());
                    break 'build_loop;
                }

                // Run the next step in the state machine
                cur_build = cur_build.clone().step(&build_tx).await.unwrap();

                if cur_build.clone().state() == BuildState::error() {
                    panic!("State machine error")
                };

                debug!("Trying to listen for output. Not don't block if nothing");
                // FIXME: https://github.com/tokio-rs/tokio/issues/3350
                //while let Ok(response) = &build_rx.try_recv() {
                while let Some(Some(response)) = &build_rx.recv().now_or_never() {
                    // TODO: Move this to be set outside the loop so we're not re-assigning so often
                    let mut build_metadata = BuildMetadata {
                        build: Some(unwrapped_request.clone()), // FIXME: We don't re-assign for the stage name
                        job_trigger: cur_build.job_trigger.into(),
                        build_state: ProtoJobState::from(cur_build.clone().state()).into(),
                        ..Default::default()
                    };

                    // Set our build ID for the client to correctly print out
                    build_metadata.id = cur_build.id.unwrap_or(0);

                    // TODO: Be more mindful about re-assigning timestamps
                    // Set timestamps
                    build_metadata.queue_time =
                        cur_build.queue_time.map(|t| prost_types::Timestamp {
                            seconds: t.timestamp(),
                            nanos: t.timestamp_subsec_nanos() as i32,
                        });

                    build_metadata.start_time =
                        cur_build.build_start_time.map(|t| prost_types::Timestamp {
                            seconds: t.timestamp(),
                            nanos: t.timestamp_subsec_nanos() as i32,
                        });

                    build_metadata.end_time =
                        cur_build.build_end_time.map(|t| prost_types::Timestamp {
                            seconds: t.timestamp(),
                            nanos: t.timestamp_subsec_nanos() as i32,
                        });

                    let mut build_record = BuildRecord {
                        build_metadata: Some(build_metadata.clone()),
                        build_output: Vec::new(),
                    };

                    //debug!("Stream OUTPUT: {:?}", response.clone().as_str());
                    let mut build_stage_output = BuildStage {
                        build_id: cur_build.id.unwrap(),
                        stage: cur_build.build_stage_name.clone(),
                        status: ProtoJobState::from(cur_build.clone().state()).into(),
                        ..Default::default()
                    };

                    build_stage_output.output = response.as_bytes().to_owned();
                    build_record.build_output.push(build_stage_output);

                    let _ = match client_tx.send(Ok(build_record.clone())).await {
                        Ok(_) => Ok(()),
                        Err(_) => Err(()),
                    };
                }
            }
        });

        Ok(Response::new(tokio_stream::wrappers::ReceiverStream::new(
            client_rx,
        )))
    }

    async fn build_stop(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        let unwrapped_request = request.into_inner();

        let orb_db = postgres::client::OrbitalDBClient::new();

        // Resolve the build number to latest if build number is 0
        let build_id = match unwrapped_request.id {
            0 => {
                if let Ok((_, repo, _)) =
                    orb_db.repo_get(&unwrapped_request.org, &unwrapped_request.git_repo)
                {
                    repo.next_build_index - 1
                } else {
                    panic!("No build id provided. Failed to query DB for latest build id")
                }
            }
            _ => unwrapped_request.id,
        };

        // Determine if build is cancelable
        match orb_db.build_summary_get(
            &unwrapped_request.org,
            &unwrapped_request.git_repo,
            &unwrapped_request.commit_hash,
            &unwrapped_request.branch,
            build_id,
        ) {
            Ok((repo, build_target, Some(summary))) => match summary.build_state {
                JobState::Queued => {
                    info!("Stop build before it even gets started");

                    // Probably change the build job state to canceled
                    let mut new_canceled_summary = summary.clone();
                    new_canceled_summary.build_state = JobState::Canceled;

                    info!("Updating build state to canceled");
                    let _build_summary_result_canceled = orb_db
                        .build_summary_update(
                            &unwrapped_request.org,
                            &repo.name,
                            &build_target.git_hash,
                            &build_target.branch,
                            build_target.build_index,
                            NewBuildSummary {
                                build_target_id: summary.build_target_id,
                                start_time: summary.start_time,
                                end_time: Some(NaiveDateTime::from_timestamp(
                                    Utc::now().timestamp(),
                                    0,
                                )),
                                build_state: postgres::schema::JobState::Canceled,
                            },
                        )
                        .expect("Unable to update build summary job state to canceled");

                    Ok(Response::new(BuildMetadata {
                        ..Default::default()
                    }))
                }
                JobState::Starting | JobState::Running => {
                    // Send build cancelation signal
                    let container_name = orbital_agent::generate_unique_build_id(
                        &unwrapped_request.org,
                        &unwrapped_request.git_repo,
                        &unwrapped_request.commit_hash,
                        &format!("{}", build_id),
                    );

                    info!("Send a cancel signal for container: {}", &container_name);

                    // Probably change the build job state to canceled
                    let rt = Runtime::new().unwrap();
                    let handle = rt.handle().clone();

                    let container_to_stop = container_name;
                    handle.spawn(async move {
                        let _ = build_engine::docker_container_stop(container_to_stop.as_ref())
                            .await
                            .expect("Sending Docker container stop failed");
                    });

                    // Update summary.build_state to JobState::Canceled
                    let mut new_canceled_summary = summary.clone();
                    new_canceled_summary.build_state = JobState::Canceled;

                    info!("Updating build state to canceled");
                    let _build_summary_result_canceled = orb_db
                        .build_summary_update(
                            &unwrapped_request.org,
                            &repo.name,
                            &build_target.git_hash,
                            &build_target.branch,
                            build_target.build_index,
                            NewBuildSummary {
                                build_target_id: summary.build_target_id,
                                start_time: summary.start_time,
                                end_time: Some(NaiveDateTime::from_timestamp(
                                    Utc::now().timestamp(),
                                    0,
                                )),
                                build_state: postgres::schema::JobState::Canceled,
                            },
                        )
                        .expect("Unable to update build summary job state to canceled");

                    Ok(Response::new(BuildMetadata {
                        ..Default::default()
                    }))
                }
                _ => {
                    println!("Build is not cancelable");
                    Err(Status::new(Code::Aborted, "Build not cancelable"))
                }
            },
            Ok((_, _, None)) => {
                // Build hasn't been queued yet
                error!("Build is not yet queued, and couldn't be canceled. This is a bug.");
                Err(Status::new(
                    Code::FailedPrecondition,
                    "FIXME: Build has not been queued yet but we can't signal a cancel",
                ))
            }
            Err(_) => {
                error!("Build was not found");
                Err(Status::new(Code::NotFound, "Build was not found"))
            }
        }
    }

    //type BuildLogsStream =
    //    Pin<Box<dyn Stream<Item = Result<BuildLogResponse, Status>> + Send + Sync + 'static>>;

    type BuildLogsStream = tokio_stream::wrappers::ReceiverStream<Result<BuildLogResponse, Status>>;

    async fn build_logs(
        &self,
        request: Request<BuildTarget>,
    ) -> Result<tonic::Response<Self::BuildLogsStream>, tonic::Status> {
        let unwrapped_request = request.into_inner();

        // Get repo id from BuildTarget
        // Connect to database. Query for the repo
        let orb_db = postgres::client::OrbitalDBClient::new();

        // Resolve the build number to latest if build number is 0
        let build_id = match unwrapped_request.id {
            0 => {
                if let Ok((_, repo, _)) =
                    orb_db.repo_get(&unwrapped_request.org, &unwrapped_request.git_repo)
                {
                    repo.next_build_index - 1
                } else {
                    panic!("No build id provided. Failed to query DB for latest build id")
                }
            }
            _ => unwrapped_request.id,
        };

        let (_repo, _build_target, build_summary_opt) = orb_db
            .build_summary_get(
                &unwrapped_request.org,
                &unwrapped_request.git_repo,
                &unwrapped_request.commit_hash,
                &unwrapped_request.branch,
                build_id,
            )
            .unwrap();

        drop(orb_db);

        let (tx, rx) = mpsc::channel(4);

        tokio::spawn(async move {
            if let Some(summary) = build_summary_opt {
                match summary.build_state {
                    JobState::Queued | JobState::Running => {
                        let container_name = orbital_agent::generate_unique_build_id(
                            &unwrapped_request.org,
                            &unwrapped_request.git_repo,
                            &unwrapped_request.commit_hash,
                            &format!("{}", build_id),
                        );

                        let mut stream =
                            build_engine::docker_container_logs(container_name.clone())
                                .await
                                .unwrap();

                        while let Some(response) = stream.recv().await {
                            let mut container_logs = BuildStage {
                                ..Default::default()
                            };

                            println!("LOGS OUTPUT: {:?}", response.clone().as_str());

                            // Adding newlines

                            let output = response.clone().as_bytes().to_owned();
                            container_logs.output = output;

                            let build_record = BuildRecord {
                                build_metadata: None,
                                build_output: vec![container_logs],
                            };

                            //
                            let build_log_response = BuildLogResponse {
                                id: build_id,
                                records: vec![build_record],
                            };

                            let _ = match tx.send(Ok(build_log_response)).await {
                                Ok(_) => Ok(()),
                                Err(mpsc::error::SendError(_)) => Err(()),
                            };
                        }
                    }

                    _ => {
                        let orb_db = postgres::client::OrbitalDBClient::new();
                        let build_stage_query = orb_db
                            .build_logs_get(
                                &unwrapped_request.org,
                                &unwrapped_request.git_repo,
                                &unwrapped_request.commit_hash,
                                &unwrapped_request.branch,
                                Some(build_id),
                            )
                            .expect("No build stages found");

                        let mut build_stage_list: Vec<orbital_headers::build_meta::BuildStage> =
                            Vec::new();
                        for (_target, _summary, stage) in build_stage_query {
                            build_stage_list.push(stage.into());
                        }

                        let build_record = BuildRecord {
                            build_metadata: None,
                            build_output: build_stage_list,
                        };

                        //
                        let build_log_response = BuildLogResponse {
                            id: build_record.build_output[0].build_id,
                            records: vec![build_record],
                        };

                        let _ = match tx.send(Ok(build_log_response)).await {
                            Ok(_) => Ok(()),
                            Err(mpsc::error::SendError(_)) => Err(()),
                        };
                    }
                }
            }
        });

        Ok(Response::new(tokio_stream::wrappers::ReceiverStream::new(
            rx,
        )))
    }

    async fn build_remove(
        &self,
        _request: Request<BuildTarget>,
    ) -> Result<Response<BuildMetadata>, Status> {
        unimplemented!();
    }

    async fn build_summary(
        &self,
        request: Request<BuildSummaryRequest>,
    ) -> Result<Response<BuildSummaryResponse>, Status> {
        let unwrapped_request = request.into_inner();
        let build_info = &unwrapped_request
            .build
            .clone()
            .expect("No build info provided");

        debug!("Received request: {:?}", &unwrapped_request);

        // Connect to database. Query for the repo
        let orb_db = postgres::client::OrbitalDBClient::new();

        let build_summary_db = orb_db
            .build_summary_list(
                &build_info.org,
                &build_info.git_repo,
                unwrapped_request.limit,
            )
            .expect("No summary returned");

        debug!("Summary: {:?}", &build_summary_db);

        let metadata_proto: Vec<BuildMetadata> = build_summary_db
            .into_iter()
            .map(|(repo, target, summary)| BuildMetadata {
                id: summary.id,
                build: Some(BuildTarget {
                    org: build_info.org.clone(),
                    git_repo: repo.name,
                    remote_uri: repo.uri,
                    branch: target.branch,
                    commit_hash: target.git_hash,
                    user_envs: match target.user_envs {
                        Some(e) => e,
                        None => "".to_string(),
                    },
                    id: target.build_index,
                    trigger: target.trigger.into(),
                    config: "".to_string(),
                }),
                job_trigger: target.trigger.into(),
                queue_time: Some(prost_types::Timestamp {
                    seconds: target.queue_time.timestamp(),
                    nanos: target.queue_time.timestamp_subsec_nanos() as i32,
                }),
                start_time: summary.start_time.map(|start_time| prost_types::Timestamp {
                    seconds: start_time.timestamp(),
                    nanos: start_time.timestamp_subsec_nanos() as i32,
                }),
                end_time: summary.end_time.map(|end_time| prost_types::Timestamp {
                    seconds: end_time.timestamp(),
                    nanos: end_time.timestamp_subsec_nanos() as i32,
                }),
                build_state: summary.build_state.into(),
            })
            .collect();

        Ok(Response::new(BuildSummaryResponse {
            summaries: metadata_proto,
        }))
    }
}
