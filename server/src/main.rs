use axum::{response::Json, routing::get, Router};

use serde_json::{json, Value};

#[tokio::main]
async fn main() {
    // build our application with a single route
    let app = Router::new()
        .route("/", get(entrypoint))
        .route("/tickets", get(tickets));

    // run our app with hyper, listening globally on port 3000
    let listener = tokio::net::TcpListener::bind("0.0.0.0:3000").await.unwrap();

    axum::serve(listener, app).await.unwrap();
}

async fn entrypoint() -> Json<Value> {
    Json(json!({"msg": "hello there!"}))
}

async fn tickets() -> Json<Value> {
    Json(json!({
        "tickets": [
            {
                "number": 1,
                "firstName": "Bilbo",
                "lastName": "Baggins",
                "entered": true
            },
            {
                "number": 2,
                "firstName": "Gandalf",
                "lastName": "Mithrandir",
                "entered": false
            },
            {
                "number": 3,
                "firstName": "Frodo",
                "lastName": "Baggins",
                "entered": false
            }
        ]
    }))
}
