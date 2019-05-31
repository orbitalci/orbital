/// This is named build_subcmd.rs bc we can't use build.rs due to overlapping with `cargo` features.

extern crate clap;

use structopt::StructOpt;

//ocelot build -acct-repo <acct>/<repo> -hash <git_hash> -branch <branch> [-latest]

#[derive(Debug, StructOpt)]
#[structopt(rename_all = "kebab_case")]
pub struct BuildOptions {
    #[structopt(long)]
    acct_repo: Option<String>,
    #[structopt(long)]
    hash : Option<String>,
    #[structopt(long)]
    branch : Option<String>,
}

// Let's define the build options here
pub fn build() {
    println!("Hello, world!");
}
