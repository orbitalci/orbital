use crate::postgres::schema::org;

#[derive(Insertable)]
#[table_name = "org"]
pub struct NewOrg {
    pub name: String,
}

#[derive(Queryable)]
pub struct Org {
    pub id: i32,
    pub name: String,
}
