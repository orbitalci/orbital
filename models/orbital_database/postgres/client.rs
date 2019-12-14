use crate::postgres::org::{NewOrg, Org};
use crate::postgres::schema::org;
use diesel::pg::PgConnection;
use diesel::prelude::*;
use log::debug;
use std::env;

pub fn establish_connection() -> PgConnection {
    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    PgConnection::establish(&database_url).expect(&format!("Error connecting to {}", database_url))
}

// FIXME: This isn't checking for existence. It is just selecting all.
pub fn new_org(conn: &PgConnection, name: &str) -> Result<Org, String> {
    // Only insert if there are no other orgs by this name
    let mut org_check: Vec<Org> = org::table
        .select(org::all_columns)
        .filter(org::name.eq(&name))
        .order_by(org::id)
        .load(conn)
        .expect("Error querying for org");

    match &org_check.len() {
        0 => {
            debug!("org doesn't exist. Inserting into db.");
            Ok(diesel::insert_into(org::table)
                .values(NewOrg {
                    name: name.to_string(),
                    ..Default::default()
                })
                .get_result(conn)
                .expect("Error saving new org"))
        }
        1 => {
            debug!("org found in db. Returning result.");
            Ok(org_check.pop().unwrap())
        }
        _ => panic!("Found more than one org by the same name in db"),
    }
}

pub fn get_org(conn: &PgConnection, name: &str) -> Result<Org, String> {
    let mut org_check: Vec<Org> = org::table
        .select(org::all_columns)
        .filter(org::name.eq(&name))
        .order_by(org::id)
        .load(conn)
        .expect("Error querying for org");

    match &org_check.len() {
        0 => {
            debug!("org doesn't exist");
            Err("Org not Found".to_string())
        }
        1 => {
            debug!("org found in db. Returning result.");
            Ok(org_check.pop().unwrap())
        }
        _ => panic!("Found more than one org by the same name in db"),
    }
}

//pub fn update_org(conn: &PgConnection, name: &str, org: Org) -> Result<Org, String> {
//    unimplemented!()
//}
//
//pub fn remove_org(conn: &PgConnection, org: Org) -> Result<(), String> {
//    unimplemented!()
//}
