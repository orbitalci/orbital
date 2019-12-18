use crate::postgres::org::{NewOrg, Org};
use crate::postgres::schema::{org, secret, ActiveState, SecretType};
use crate::postgres::secret::{NewSecret, Secret};
use agent_runtime::vault::orb_vault_path;
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
    let org_check: Result<Org, _> = org::table
        .select(org::all_columns)
        .filter(org::name.eq(&name))
        .order_by(org::id)
        .get_result(conn);

    match org_check {
        Err(_e) => {
            debug!("org doesn't exist. Inserting into db.");
            Ok(diesel::insert_into(org::table)
                .values(NewOrg {
                    name: name.to_string(),
                    ..Default::default()
                })
                .get_result(conn)
                .expect("Error saving new org"))
        }
        Ok(o) => {
            debug!("org found in db. Returning result.");
            Ok(o)
        }
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
    let org = org_get(conn, org).expect("Unable to find org");

    let secret_check: Result<Secret, _> = secret::table
        .select(secret::all_columns)
        .filter(secret::name.eq(&name))
        .order_by(secret::id)
        .get_result(conn);

    match secret_check {
        Err(_e) => {
            debug!("secret doesn't exist. Inserting into db.");
            Ok(diesel::insert_into(secret::table)
                .values(NewSecret {
                    name: name.to_string(),
                    org_id: org.id,
                    secret_type: secret_type,
                    vault_path: orb_vault_path(
                        &org.name,
                        name,
                        format!("{:?}", &secret_type).as_str(),
                    ),
                    ..Default::default()
                })
                .get_result(conn)
                .expect("Error saving new secret"))
        }
        Ok(s) => {
            debug!("secret found in db. Returning result.");
            Ok(s)
        }
    }
}

pub fn secret_get(
    conn: &PgConnection,
    org: &str,
    name: &str,
    secret_type: SecretType,
) -> Result<Secret, String> {
    let org_db = org_get(conn, org).expect("Unable to find org");

    let secret: Secret = secret::table
        .select(secret::all_columns)
        .filter(secret::org_id.eq(&org_db.id))
        .filter(secret::name.eq(&name))
        .order_by(secret::id)
        .get_result(conn)
        .expect("Error querying for secret");

    Ok(secret)
}

pub fn secret_update(
    conn: &PgConnection,
    org: &str,
    name: &str,
    update_secret: NewSecret,
) -> Result<Secret, String> {
    let org_db = org_get(conn, org).expect("Unable to find org");

    let secret_update: Secret = diesel::update(secret::table)
        .filter(secret::org_id.eq(&org_db.id))
        .filter(secret::name.eq(&name))
        .set(update_secret)
        .get_result(conn)
        .expect("Error updating secret");

    debug!("Result after update: {:?}", &secret_update);

    Ok(secret_update)
}

pub fn secret_remove(
    conn: &PgConnection,
    org: &str,
    name: &str,
    secret_type: SecretType,
) -> Result<Secret, String> {
    let org_db = org_get(conn, org).expect("Unable to find org");

    let secret_delete: Secret = diesel::delete(secret::table)
        .filter(secret::org_id.eq(&org_db.id))
        .filter(secret::name.eq(&name))
        .get_result(conn)
        .expect("Error deleting secret");

    Ok(secret_delete)
}

pub fn secret_list(
    conn: &PgConnection,
    org: &str,
    filter: Option<SecretType>,
) -> Result<Vec<Secret>, String> {
    let org_db = org_get(conn, org).expect("Unable to find org");

    let query: Vec<Secret> = match filter {
        None => secret::table
            .select(secret::all_columns)
            .filter(secret::org_id.eq(&org_db.id))
            .order_by(secret::id)
            .load(conn)
            .expect("Error getting all secret records"),
        Some(f) => secret::table
            .select(secret::all_columns)
            .filter(secret::org_id.eq(&org_db.id))
            //.filter(secret::secret_type.eq(SecretType::from(f))) // Not working yet.
            .order_by(secret::id)
            .load(conn)
            .expect("Error getting secret records by filter"),
    };

    Ok(query)
}
