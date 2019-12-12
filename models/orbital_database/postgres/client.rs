use crate::postgres::org::{NewOrg, Org};
use crate::postgres::schema::org;
use diesel::pg::PgConnection;
use diesel::prelude::*;
use std::env;

pub fn establish_connection() -> PgConnection {
    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    PgConnection::establish(&database_url).expect(&format!("Error connecting to {}", database_url))
}

pub fn new_org(conn: &PgConnection, org_form: NewOrg) -> Org {
    // TODO: Only insert if there are no other orgs by this name

    let mut org_check: Vec<Org> = org::table
        .select(org::all_columns)
        .order_by(org::id)
        .load(conn)
        .expect("Error querying for org");

    match &org_check.len() {
        0 => diesel::insert_into(org::table)
            .values(&org_form)
            .get_result(conn)
            .expect("Error saving new org"),
        1 => org_check.pop().unwrap(),
        _ => panic!("Found more than one org by the same name in db"),
    }
}
