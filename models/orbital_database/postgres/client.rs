use crate::postgres::org::{NewOrg, Org};
use crate::postgres::schema::{org, secret, ActiveState, SecretType};
use crate::postgres::secret::{NewSecret, Secret};
use diesel::pg::PgConnection;
use diesel::prelude::*;
use log::debug;
use std::env;

pub fn establish_connection() -> PgConnection {
    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL must be set");
    PgConnection::establish(&database_url).expect(&format!("Error connecting to {}", database_url))
}

pub fn org_add(conn: &PgConnection, name: &str) -> Result<Org, String> {
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

pub fn org_get(conn: &PgConnection, name: &str) -> Result<Org, String> {
    let mut org_check: Vec<Org> = org::table
        .select(org::all_columns)
        .filter(org::name.eq(&name))
        .order_by(org::id)
        .load(conn)
        .expect("Error querying for org");

    match &org_check.len() {
        0 => {
            debug!("org not found in db");
            Err("Org not Found".to_string())
        }
        1 => {
            debug!("org found in db. Returning result.");
            Ok(org_check.pop().unwrap())
        }
        _ => panic!("Found more than one org by the same name in db"),
    }
}

pub fn org_update(conn: &PgConnection, name: &str, update_org: NewOrg) -> Result<Org, String> {
    let org_update: Org = diesel::update(org::table)
        .filter(org::name.eq(&name))
        .set(update_org)
        .get_result(conn)
        .expect("Error updating org");

    debug!("Result after update: {:?}", &org_update);

    Ok(org_update)
}

pub fn org_remove(conn: &PgConnection, name: &str) -> Result<Org, String> {
    let org_delete: Org = diesel::delete(org::table)
        .filter(org::name.eq(&name))
        .get_result(conn)
        .expect("Error deleting org");

    Ok(org_delete)
}

pub fn org_list(conn: &PgConnection) -> Result<Vec<Org>, String> {
    let query: Vec<Org> = org::table
        .select(org::all_columns)
        .order_by(org::id)
        .load(conn)
        .expect("Error getting all order records");

    Ok(query)
}

pub fn secret_add(
    conn: &PgConnection,
    org: &str,
    name: &str,
    secret_type: SecretType,
) -> Result<Secret, String> {
    //    string org = 1;
    //    string name = 2;
    //    orbital_types.SecretType secret_type = 3;
    //
    //    vault_path
    //    active_state

    //let mut secret_check: Vec<Secret> = secret::table
    //    .select(secret::all_columns)
    //    .filter(secret::name.eq(&name))
    //    .order_by(secret::id)
    //    .load(conn)
    //    .expect("Error querying for secret");

    //let mut secret_check: Vec<Secret> = secret::table
    //    .select(secret::all_columns)
    //    .load(conn)
    //    .expect("Error querying for secret");

    //match &secret_check.len() {
    //    0 => {
    //        debug!("secret doesn't exist. Inserting into db.");
    //        Ok(diesel::insert_into(secret::table)
    //            .values(NewSecret {
    //                name: name.to_string(),
    //                ..Default::default()
    //            })
    //            .get_result(conn)
    //            .expect("Error saving new secret"))
    //    }
    //    1 => {
    //        debug!("secret found in db. Returning result.");
    //        Ok(secret_check.pop().unwrap())
    //    }
    //    _ => panic!("Found more than one secret in the org by the same name in db"),
    //}

    // This works
    diesel::insert_into(secret::table)
        .values(vec![NewSecret {
            name: name.to_string(),
            org_id: 1,
            secret_type: SecretType::SshKey.into(),
            ..Default::default()
        }])
        .execute(conn)
        .expect("Error inserting secret into db");

    Ok(Secret {
        ..Default::default()
    })
}
