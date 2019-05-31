extern crate structopt;
use structopt::StructOpt;

#[derive(Debug, StructOpt, Copy, Clone)]
#[structopt(rename_all = "kebab_case")]
pub enum ResourceAction {
    Add,
    Delete,
    List,
}