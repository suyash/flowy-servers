#![feature(plugin)]
#![plugin(rocket_codegen)]

extern crate redis;
extern crate rocket;

#[macro_use] extern crate rocket_contrib;
#[macro_use] extern crate serde_derive;

use std::convert::From;
use std::io::Cursor;

use redis::{Client, Commands, RedisError};

use rocket::{Outcome, Request, Response, Rocket, State};
use rocket::fairing::AdHoc;
use rocket::http::{Header, ContentType, Method, Status};
use rocket::request::{self, FromRequest};
use rocket_contrib::{Json, Value};

fn main() {
    let client = Client::open("redis://redis:6379").unwrap();
    rocket(client).launch();
}

fn rocket(client: Client) -> Rocket {
    rocket::ignite()
        .mount("/", routes![set, get, del])
        .attach(AdHoc::on_response(|request: &Request, response: &mut Response| {
            // https://github.com/SergioBenitez/Rocket/issues/25#issuecomment-313895086

            if request.method() == Method::Options || response.content_type() == Some(ContentType::JSON) {
                response.set_header(Header::new("Access-Control-Allow-Origin", "*"));
                response.set_header(Header::new("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE"));
                response.set_header(Header::new("Access-Control-Allow-Headers", "Content-Type, Accept, X-API-Key"));
                response.set_header(Header::new("Access-Control-Allow-Credentials", "true"));
            }

            if request.method() == Method::Options {
                response.set_header(ContentType::Plain);
                response.set_sized_body(Cursor::new(""));
                response.set_status(Status::Ok);
            }
        }))
        .manage(client)
}

#[derive(Serialize, Deserialize)]
struct Task {
    id: String,
    text: String,
    checked: bool,
    children: Vec<String>
}

/// set will create or update a new task
#[post("/set", format = "application/json", data = "<task>")]
fn set(task: Json<Task>, _key: ApiKey, client: State<Client>) -> Result<Json<Value>, RedisError> {
    let conn = client.get_connection()?;

    conn.hset(task.0.id.clone(), "text", task.0.text)?;
    conn.hset(task.0.id.clone(), "checked", task.0.checked)?;

    let children_id = task.0.id.clone() + "_children";
    conn.del(children_id.clone())?;
    for child in task.0.children {
        conn.lpush(children_id.clone(), child)?;
    }

    Ok(Json(json!({
        "ok": true,
    })))
}

/// get will get a task if it exists, or it will 500 out
#[get("/<id>", format = "application/json")]
fn get(id: String, _key: ApiKey, client: State<Client>) -> Result<Json<Task>, RedisError> {
    let conn = client.get_connection()?;

    let checked: String = conn.hget(id.clone(), "checked")?;

    Ok(Json(Task{
        id: id.clone(),
        text: conn.hget(id.clone(), "text")?,
        checked: checked == "true",
        children: conn.lrange(id + "_children", 0, -1)?,
    }))
}

/// delete will remove a task
#[delete("/<id>", format = "application/json")]
fn del(id: String, _key: ApiKey, client: State<Client>) -> Result<Json<Value>, RedisError> {
    let conn = client.get_connection()?;
    conn.del(id)?;

    Ok(Json(json!({
        "ok": true,
    })))
}

const API_KEY: &'static str = env!("API_KEY");

/// https://api.rocket.rs/rocket/request/trait.FromRequest.html#example-1
struct ApiKey(String);

impl<'a, 'r> FromRequest<'a, 'r> for ApiKey {
    type Error = ();

    fn from_request(request: &'a Request<'r>) -> request::Outcome<ApiKey, ()> {
        let keys: Vec<&str> = request.headers().get("X-API-Key").collect();
        if keys.len() != 1 {
            return Outcome::Failure((Status::BadRequest, ()));
        }

        let key = keys[0];
        if key != API_KEY {
            Outcome::Forward(())
        } else {
            Outcome::Success(ApiKey(key.to_owned()))
        }
    }
}

#[cfg(test)]
mod test {
    extern crate serde_json;

    use rocket::local::Client;
    use rocket::http::{ContentType, Status};

    use super::{rocket, Task};

    #[test]
    fn works() {
        let client = redis::Client::open("redis://127.0.0.1:6379").unwrap();
        let client = Client::new(rocket(client)).unwrap();

        let task = Task{
            id: "test".to_owned(),
            text: "testtest".to_owned(),
            checked: true,
            children: vec![],
        };

        let res = client
            .post("/set")
            .header(ContentType::JSON)
            .body(serde_json::to_string(&task).unwrap())
            .dispatch();

        assert_eq!(res.status(), Status::Ok);

        let mut res = client
            .get(format!("/{}", task.id))
            .header(ContentType::JSON)
            .dispatch();

        assert_eq!(res.status(), Status::Ok);

        let s = res.body().unwrap().into_string().unwrap();
        let target: Task = serde_json::from_str(&s).unwrap();

        assert_eq!(target.id, task.id);
        assert_eq!(target.text, task.text);
        assert_eq!(target.checked, task.checked);
        assert_eq!(target.children, task.children);

        let res = client
            .delete(format!("/{}", task.id))
            .header(ContentType::JSON)
            .dispatch();

        assert_eq!(res.status(), Status::Ok);
    }
}
