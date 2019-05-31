pub mod apple;
pub mod env;
pub mod helm;
pub mod k8s;
pub mod slack;
pub mod artifact;
pub mod ssh;
pub mod vcs;

use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub enum CredType {
    Apple(apple::SubOption),
    Env(env::SubOption),
    Helm(helm::SubOption),
    K8s(k8s::SubOption),
    Slack(slack::SubOption),
    Artifact(artifact::SubOption),
    Ssh(ssh::SubOption),
    Vcs(vcs::SubOption),
}

// Handle the command line control flow
pub fn subcommand_handler(args: &CredType) {
    println!("Placeholder for handling creds");
    println!("{:?}", args);

    match args {
        CredType::Apple(a) => apple::subcommand_handler(a),
        CredType::Env(a) => env::subcommand_handler(a),
        CredType::Helm(a) => helm::subcommand_handler(a),
        CredType::K8s(a) => k8s::subcommand_handler(a),
        CredType::Slack(a) => slack::subcommand_handler(a),
        CredType::Artifact(a) => artifact::subcommand_handler(a),
        CredType::Ssh(a) => ssh::subcommand_handler(a),
        CredType::Vcs(a) => vcs::subcommand_handler(a),
    }
}